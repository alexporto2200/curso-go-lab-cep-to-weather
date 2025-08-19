package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro"`
}

type NominatimResponse struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type WeatherResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	r := gin.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Iniciando API CEP to Weather na porta %s\n", port)
	fmt.Printf("Endpoint disponível: http://localhost:%s/{cep}\n", port)

	r.GET("/:cep", getWeather)

	fmt.Printf("Servidor iniciado com sucesso!\n")
	r.Run(":" + port)
}

func getWeather(c *gin.Context) {
	cep := c.Param("cep")

	fmt.Printf("Consultando CEP: %s\n", cep)

	// Validar formato do CEP (8 dígitos)
	if !isValidCEP(cep) {
		fmt.Printf("CEP inválido: %s\n", cep)
		c.JSON(422, gin.H{"error": "invalid zipcode"})
		return
	}

	// Consultar CEP
	location, err := getLocationFromCEP(cep)
	if err != nil {
		fmt.Printf("CEP não encontrado: %s - Erro: %v\n", cep, err)
		c.JSON(404, gin.H{"error": "can not find zipcode"})
		return
	}

	fmt.Printf("Localização encontrada: %s\n", location)

	// Obter coordenadas
	lat, lon, err := getCoordinates(location)
	if err != nil {
		fmt.Printf("Erro ao obter coordenadas para %s: %v\n", location, err)
		// Continua sem coordenadas, usa localização
	} else {
		fmt.Printf("Coordenadas: %.6f, %.6f\n", lat, lon)
	}

	// Consultar temperatura
	// temperature, err := getTemperature(location)
	// if err != nil {
	// 	fmt.Printf("Erro ao obter temperatura para %s: %v\n", location, err)
	// 	c.JSON(500, gin.H{"error": "error getting temperature, " + err.Error()})
	// 	return
	// }

	temperatureFromCoordinates, err := getTemperatureFromCoordinates(lat, lon, location)
	if err != nil {
		fmt.Printf("Erro ao obter temperatura para %s: %v\n", location, err)
		c.JSON(500, gin.H{"error": "error getting temperature, " + err.Error()})
		return
	}

	fmt.Printf("Temperatura de %s: %.1f°C\n", location, temperatureFromCoordinates)

	// Converter temperaturas
	tempC := temperatureFromCoordinates
	tempF := celsiusToFahrenheit(tempC)
	tempK := celsiusToKelvin(tempC)

	fmt.Printf("Temperaturas obtidas - C: %.1f°C, F: %.1f°F, K: %.1fK\n", tempC, tempF, tempK)

	response := TemperatureResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	c.JSON(200, response)
}

func isValidCEP(cep string) bool {
	// Remove caracteres não numéricos
	re := regexp.MustCompile(`[^0-9]`)
	cleanCEP := re.ReplaceAllString(cep, "")

	// Verifica se tem exatamente 8 dígitos
	return len(cleanCEP) == 8
}

func getLocationFromCEP(cep string) (string, error) {
	// Remove caracteres não numéricos
	re := regexp.MustCompile(`[^0-9]`)
	cleanCEP := re.ReplaceAllString(cep, "")

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cleanCEP)
	fmt.Printf("Consultando ViaCEP: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro na requisição ViaCEP: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler resposta ViaCEP: %v\n", err)
		return "", err
	}

	var viaCEPResp ViaCEPResponse
	err = json.Unmarshal(body, &viaCEPResp)
	if err != nil {
		fmt.Printf("Erro ao decodificar JSON ViaCEP: %v\n", err)
		return "", err
	}

	if viaCEPResp.Erro {
		fmt.Printf("CEP não encontrado no ViaCEP: %s\n", cleanCEP)
		return "", fmt.Errorf("CEP não encontrado")
	}

	fmt.Printf("ViaCEP retornou: %s, %s\n", viaCEPResp.Localidade, viaCEPResp.Uf)

	// Retorna cidade, estado
	return fmt.Sprintf("%s,%s", viaCEPResp.Localidade, viaCEPResp.Uf), nil
}

func getCoordinates(location string) (float64, float64, error) {
	locationEncoded := url.QueryEscape(location)
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s,Brazil&format=json&limit=1", locationEncoded)

	fmt.Printf("Consultando coordenadas: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro na requisição de coordenadas: %v\n", err)
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler resposta de coordenadas: %v\n", err)
		return 0, 0, err
	}

	var nominatimResp []NominatimResponse
	err = json.Unmarshal(body, &nominatimResp)
	if err != nil {
		fmt.Printf("Erro ao decodificar JSON de coordenadas: %v\n", err)
		return 0, 0, err
	}

	if len(nominatimResp) == 0 {
		fmt.Printf("Coordenadas não encontradas para: %s\n", location)
		return 0, 0, fmt.Errorf("coordenadas não encontradas")
	}

	lat, err := strconv.ParseFloat(nominatimResp[0].Lat, 64)
	if err != nil {
		fmt.Printf("Erro ao converter latitude: %v\n", err)
		return 0, 0, err
	}

	lon, err := strconv.ParseFloat(nominatimResp[0].Lon, 64)
	if err != nil {
		fmt.Printf("Erro ao converter longitude: %v\n", err)
		return 0, 0, err
	}

	fmt.Printf("Coordenadas obtidas: %.6f, %.6f para %s\n", lat, lon, location)

	return lat, lon, nil
}

// isso não funciona pra minha cidade hehehe
// func getTemperature(location string) (float64, error) {
// 	// Usando WeatherAPI
// 	apiKey := os.Getenv("WEATHER_API")
// 	if apiKey == "" {
// 		fmt.Printf("WEATHER_API não configurada\n")
// 		return 0, fmt.Errorf("WEATHER_API environment variable not set")
// 	}

// 	locationEncoded := url.QueryEscape(location)

// 	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, locationEncoded)
// 	fmt.Printf("Consultando WeatherAPI:q=%s\n", location)

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		fmt.Printf("Erro na requisição WeatherAPI: %v\n", err)
// 		return 0, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Printf("Erro ao ler resposta WeatherAPI: %v\n", err)
// 		return 0, err
// 	}

// 	var weatherResp WeatherResponse
// 	err = json.Unmarshal(body, &weatherResp)
// 	if err != nil {
// 		fmt.Printf("Erro ao decodificar JSON WeatherAPI: %v\n", err)
// 		return 0, err
// 	}

// 	fmt.Printf("WeatherAPI retornou temperatura: %.1f°C para %s\n", weatherResp.Current.TempC, location)

// 	return weatherResp.Current.TempC, nil
// }

func getTemperatureFromCoordinates(lat float64, lon float64, location string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API")
	if apiKey == "" {
		fmt.Printf("WEATHER_API não configurada\n")
		return 0, fmt.Errorf("WEATHER_API environment variable not set")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%f,%f&aqi=no", apiKey, lat, lon)
	fmt.Printf("Consultando WeatherAPI:q=%f,%f\n", lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Erro na requisição WeatherAPI: %v\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler resposta WeatherAPI: %v\n", err)
		return 0, err
	}

	var weatherResp WeatherResponse
	err = json.Unmarshal(body, &weatherResp)
	if err != nil {
		fmt.Printf("Erro ao decodificar JSON WeatherAPI: %v\n", err)
		return 0, err
	}

	fmt.Printf("WeatherAPI retornou temperatura: %.1f°C para %s, %f,%f\n", weatherResp.Current.TempC, location, lat, lon)

	return weatherResp.Current.TempC, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
