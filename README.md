# Network Limiter for Azure PaaS

## Overview

![](/logo.svg)

Network Limiter for Azure PaaS (aka nlap) is CLI tool, written on Golang, that limits network access to Azure PaaS (Platform-as-a-Service) instances. Under the hood it uses [Azure Go SDK](https://github.com/Azure/azure-sdk-for-go).

## Quick start

To check available commands run the tool with -h flag:
```
./nlap -h
```

By default, CLI do not overwrite exsisting rules (if there are any), but append them. As source, for whitelisting, could be used list of IPs (separated by semicolon) from CLI or/and external URLs with allowed IPs(supports 'https' only). 

### Examples

Add to allowed IPs a list stored in a url (appends only):
```
./nlap set -u "https://raw.githubusercontent.com/groovy-sky/azure-ip-ranges/main/ip/ApiManagement.WestEurope.txt" -s "/subscriptions/<sub-id>/resourceGroups/<res-grp>/Microsoft.Storage/storageAccounts/<res-name>"
```

Allow to access storage accounts from certain IPs only (exsisting rules will be removed) with enhanced security (Minimum TLS version 1.2, no anonymous access to blob containers allowed, HTTPS access only):
```
./nlap set -i "1.1.1.1;2.2.2.2" -s "/subscriptions/<sub-id-1>/resourceGroups/<res-grp-1>/Microsoft.Storage/storageAccounts/<res-name-1>;/subscriptions/<sub-id-2>/resourceGroups/<res-grp-2>/Microsoft.Storage/storageAccounts/<res-name-2>" -e -f
```

Fully disable access (if you planning to use Private Endpoints only):
```
./nlap set -s "/subscriptions/<sub-id>/resourceGroups/<res-grp>/Microsoft.Storage/storageAccounts/<res-name>" -f 
```

## ToDo
- [x] Check how it works for V1 Storage
- [x] Add posibility to get inputs from web
- [] Develop Azure Function, which would trigger by timer and blob modification
- [x] Implement goroutine for parallel exec
- [x] Implement force
- [x] Implement secure mode - force use https only, denies public access etc.
- [x] Change CLI lib
- [] Add Windows OS for build
- [] Add disable public access option with exsisting rules cleanup
- [] Add another PaaS service support

## Related materials

https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity

https://learn.microsoft.com/en-us/rest/api/

https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/examples_test.go

https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/main/sdk/resourcemanager/resource/resources/main.go

https://learn.microsoft.com/en-us/rest/api/storagerp/storage-accounts/get-properties?tabs=Go#storageaccountgetproperties

https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage#section-readme