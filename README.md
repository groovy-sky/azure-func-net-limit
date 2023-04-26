# Network Limiter for Azure PaaS

## Overview

Network Limiter for Azure PaaS is CLI tool, written on Golang, that limits network access to Azure PaaS (Platform-as-a-Service) instances. Under the hood it uses [Azure Go SDK](https://github.com/Azure/azure-sdk-for-go).

## Quick start

## ToDo
- [x] Check how it works for V1 Storage
- [x] Add posibility to get inputs from web
- [] Develop Azure Function, which would trigger by timer and blob modification
- [x] Implement goroutine for parallel exec

## Related materials

https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity

https://learn.microsoft.com/en-us/rest/api/

https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/examples_test.go

https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/main/sdk/resourcemanager/resource/resources/main.go

https://learn.microsoft.com/en-us/rest/api/storagerp/storage-accounts/get-properties?tabs=Go#storageaccountgetproperties

https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage#section-readme