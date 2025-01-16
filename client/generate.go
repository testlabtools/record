package client

//go:generate curl -sS --retry 3 https://eu.testlab.tools/api.yml -o api.yml
//go:generate oapi-codegen -config cfg.yml api.yml
