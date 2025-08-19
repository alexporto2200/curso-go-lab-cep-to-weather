#!/bin/bash

sudo apt-get update
sudo apt-get install apt-transport-https ca-certificates gnupg curl


curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
    | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list

sudo apt-get update && sudo apt-get install google-cloud-cli


# outras dependencias
sudo apt-get install google-cloud-cli google-cloud-cli google-cloud-cli-gke-gcloud-auth-plugin\
 kubectl google-cloud-cli-skaffold google-cloud-cli-minikube google-cloud-cli-log-streaming


##########

# login
gcloud auth login

gcloud config set project alexpessoal

gcloud projects list


# deploy
gcloud run deploy curso-go-lab-cep-to-weather \
  --project alexpessoal \
  --image gcr.io/alexpessoal/curso-go-lab-cep-to-weather \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --cpu 1 \
  --memory 256Mi \
  --concurrency 80 \
  --timeout 300 \
  --set-env-vars WEATHER_API="$(cat .env | grep WEATHER_API | cut -d '=' -f2)"

# ver logs
gcloud beta run services logs tail curso-go-lab-cep-to-weather --project alexpessoal --region us-central1 

# ver instancia status / url
gcloud run services describe curso-go-lab-cep-to-weather --project alexpessoal --region us-central1 