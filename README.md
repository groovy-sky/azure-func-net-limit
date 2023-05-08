# Network Limiter for Azure PaaS

## Overview

![](/logo.svg)

Network Limiter for Azure PaaS (aka nlap) is CLI tool, written on Golang, that limits network access to Azure PaaS (Platform-as-a-Service) instances. Under the hood it uses [Azure Go SDK](https://github.com/Azure/azure-sdk-for-go).

## Quick start

### Installation

To build from scratch you'll need Go >= 1.19. Open terminal and execute following command:

```
export GOPATH="$HOME/go"
PATH="$GOPATH/bin:$PATH"
go install github.com/groovy-sky/nlap/v2@latest
```

Another way how you can get this tool - check the latest version under [releases section](/releases)

### Examples

To check available commands run the tool with -h flag:
```
./nlap -h
```

By default, CLI do not overwrite exsisting rules (if there are any), but append them. As source, for whitelisting, could be used list of IPs (separated by semicolon) from CLI or/and external URLs with allowed IPs(supports 'https' only). 

Add to allowed IPs a list stored in URL (appends only):
```
./nlap set -u "https://raw.githubusercontent.com/groovy-sky/azure-ip-ranges/main/ip/ApiManagement.WestEurope.txt" -s "/subscriptions/<sub-id>/resourceGroups/<res-grp>/Microsoft.Storage/storageAccounts/<res-name>"
```

Allow to access storage accounts from certain IPs only (exsisting rules will be removed):
```
./nlap set -i "1.1.1.1;2.2.2.2" -s "/subscriptions/<sub-id-1>/resourceGroups/<res-grp-1>/Microsoft.Storage/storageAccounts/<res-name-1>;/subscriptions/<sub-id-2>/resourceGroups/<res-grp-2>/Microsoft.Storage/storageAccounts/<res-name-2>" -f
```
Append access with current environment public IP (using external service for showing IP) and enable enhanced security (setup Minimum TLS version to 1.2, no anonymous access to blob containers will be allowed, HTTPS access accepted only):

```
./nlap set -u "https://api.ipify.org" -s "/subscriptions/<sub-id>/resourceGroups/<res-grp>/Microsoft.Storage/storageAccounts/<res-name>" -e
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
- [] Add what is my IP funcionality

## Related materials

https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity

https://learn.microsoft.com/en-us/rest/api/

https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/examples_test.go

https://github.com/Azure-Samples/azure-sdk-for-go-samples/blob/main/sdk/resourcemanager/resource/resources/main.go

https://learn.microsoft.com/en-us/rest/api/storagerp/storage-accounts/get-properties?tabs=Go#storageaccountgetproperties

https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage#section-readme