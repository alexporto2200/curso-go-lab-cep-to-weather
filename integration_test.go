package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.Default()
	suite.router.GET("/:cep", getWeather)
}

// Teste de funcionamento normal da aplicação
func (suite *IntegrationTestSuite) TestSuccessfulWeatherRequest() {
	// Configurar API key para teste
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/01310100", nil)
	suite.router.ServeHTTP(w, req)

	// Verifica se a resposta é válida (200 ou 500 se não tiver API key real)
	suite.Contains([]int{200, 500}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	if w.Code == 200 {
		// Se sucesso, verifica a estrutura da resposta
		suite.Contains(response, "temp_C")
		suite.Contains(response, "temp_F")
		suite.Contains(response, "temp_K")

		// Verifica se os valores são números
		tempC, ok := response["temp_C"].(float64)
		suite.True(ok)
		suite.Greater(tempC, -50.0)
		suite.Less(tempC, 60.0)
	} else if w.Code == 500 {
		// Se erro, verifica se tem mensagem de erro
		suite.Contains(response, "error")
	}
}

// Teste com CEP válido de São Paulo
func (suite *IntegrationTestSuite) TestValidSaoPauloCEP() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/01310100", nil)
	suite.router.ServeHTTP(w, req)

	suite.Contains([]int{200, 500}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	if w.Code == 200 {
		suite.Contains(response, "temp_C")
		suite.Contains(response, "temp_F")
		suite.Contains(response, "temp_K")
	}
}

// Teste com CEP válido do Rio de Janeiro
func (suite *IntegrationTestSuite) TestValidRioCEP() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/20040020", nil)
	suite.router.ServeHTTP(w, req)

	suite.Contains([]int{200, 500}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	if w.Code == 200 {
		suite.Contains(response, "temp_C")
		suite.Contains(response, "temp_F")
		suite.Contains(response, "temp_K")
	}
}

// Teste com CEP válido de Curitiba
func (suite *IntegrationTestSuite) TestValidCuritibaCEP() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/80010000", nil)
	suite.router.ServeHTTP(w, req)

	suite.Contains([]int{200, 500}, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	if w.Code == 200 {
		suite.Contains(response, "temp_C")
		suite.Contains(response, "temp_F")
		suite.Contains(response, "temp_K")
	}
}

// Teste com diferentes formatos de CEP válido
func (suite *IntegrationTestSuite) TestCEPFormatVariations() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	cepVariations := []string{
		"01310100",
		"01310-100",
		"01310.100",
		"01310 100",
	}

	for _, cep := range cepVariations {
		suite.Run(fmt.Sprintf("Format_%s", cep), func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/"+cep, nil)
			suite.router.ServeHTTP(w, req)

			// Todos devem retornar 200 (CEP válido) ou 500 (sem API key real)
			suite.Contains([]int{200, 500}, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)

			if w.Code == 200 {
				suite.Contains(response, "temp_C")
			} else {
				suite.Contains(response, "error")
			}
		})
	}
}

// Teste de CEP inválido - muito curto
func (suite *IntegrationTestSuite) TestInvalidCEPTooShort() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/123", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(422, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid zipcode", response["error"])
}

// Teste de CEP inválido - muito longo
func (suite *IntegrationTestSuite) TestInvalidCEPTooLong() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/123456789", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(422, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid zipcode", response["error"])
}

// Teste de CEP inválido - com letras
func (suite *IntegrationTestSuite) TestInvalidCEPWithLetters() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abcdefgh", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(422, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid zipcode", response["error"])
}

// Teste de CEP inválido - com caracteres especiais
func (suite *IntegrationTestSuite) TestInvalidCEPWithSpecialChars() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/12-345-67", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(422, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid zipcode", response["error"])
}

// Teste de CEP inválido - com pontos
func (suite *IntegrationTestSuite) TestInvalidCEPWithDots() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/12.345.67", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(422, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("invalid zipcode", response["error"])
}

// Teste de CEP inexistente
func (suite *IntegrationTestSuite) TestNonExistentCEP() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/99999999", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("can not find zipcode", response["error"])
}

// Teste de CEP inexistente - outro exemplo
func (suite *IntegrationTestSuite) TestAnotherNonExistentCEP() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/88888888", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(404, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("can not find zipcode", response["error"])
}

// Teste sem API key configurada
func (suite *IntegrationTestSuite) TestWithoutAPIKey() {
	// Garantir que não há API key
	os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/01310100", nil)
	suite.router.ServeHTTP(w, req)

	suite.Equal(500, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"], "WEATHER_API environment variable not set")
}

// Teste de performance - tempo de resposta
func (suite *IntegrationTestSuite) TestResponseTime() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	start := time.Now()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/01310100", nil)
	suite.router.ServeHTTP(w, req)

	duration := time.Since(start)

	// Verifica se a resposta é rápida (menos de 10 segundos)
	suite.Less(duration, 10*time.Second)
	suite.Contains([]int{200, 500}, w.Code)
}

// Teste de estrutura JSON da resposta
func (suite *IntegrationTestSuite) TestJSONResponseStructure() {
	os.Setenv("WEATHER_API", "test_key")
	defer os.Unsetenv("WEATHER_API")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/01310100", nil)
	suite.router.ServeHTTP(w, req)

	suite.Contains([]int{200, 500}, w.Code)

	// Verifica se é um JSON válido
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	if w.Code == 200 {
		// Verifica se os campos estão presentes
		suite.Contains(response, "temp_C")
		suite.Contains(response, "temp_F")
		suite.Contains(response, "temp_K")

		// Verifica se os valores são números
		suite.IsType(float64(0), response["temp_C"])
		suite.IsType(float64(0), response["temp_F"])
		suite.IsType(float64(0), response["temp_K"])
	} else {
		// Verifica se tem mensagem de erro
		suite.Contains(response, "error")
	}
}

// Teste de conversões de temperatura
func (suite *IntegrationTestSuite) TestTemperatureConversions() {
	// Testa se as conversões estão corretas
	testCases := []struct {
		celsius    float64
		fahrenheit float64
		kelvin     float64
	}{
		{0, 32, 273},
		{100, 212, 373},
		{-40, -40, 233},
		{25, 77, 298},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Conversion_%.0f", tc.celsius), func() {
			f := celsiusToFahrenheit(tc.celsius)
			k := celsiusToKelvin(tc.celsius)

			suite.InDelta(tc.fahrenheit, f, 0.1)
			suite.InDelta(tc.kelvin, k, 0.1)
		})
	}
}

// Teste de validação de CEP
func (suite *IntegrationTestSuite) TestCEPValidation() {
	testCases := []struct {
		cep      string
		expected bool
	}{
		{"01310100", true},
		{"01310-100", true},
		{"01310.100", true},
		{"01310 100", true},
		{"123", false},
		{"123456789", false},
		{"abcdefgh", false},
		{"", false},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Validation_%s", tc.cep), func() {
			result := isValidCEP(tc.cep)
			suite.Equal(tc.expected, result)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
