# Nightfall Go SDK #

nightfall-go-sdk is a Go client library for accessing the Nightfall API. 
It allows you to add functionality to your applications to
scan plain text and files in order to detect different categories of information. You can leverage any of
the detectors in Nightfall's pre-built library, or you may programmatically define your own custom detectors. 

Additionally, this library provides convenient features including a streamlined function to manage the multi-stage file upload process.

To obtain an API Key, login to the [Nightfall dashboard](https://app.nightfall.ai/) and click the section
titled "Manage API Keys".

See our [developer documentation](https://docs.nightfall.ai/docs/entities-and-terms-to-know) for more details about
integrating with the Nightfall API.

## Installation ##

Nightfall Go SDK is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/nightfallai/nightfall-go-sdk
```

will resolve and add the package to the current development module, along with its dependencies.

## Usage

### Scanning Plain Text

Nightfall provides pre-built detector types, covering data types ranging from PII to PHI to credentials. The following
snippet shows an example of how to scan using pre-built detectors.

####  Sample Code
See [examples/text/text\_scanner.go](examples/text/text_scanner.go) for an example

### Scanning Files

Scanning common file types like PDFs or office documents typically requires cumbersome text
extraction methods like OCR.

Rather than implementing this functionality yourself, the Nightfall API allows you to upload the
original files, and then we'll handle the heavy lifting.

The file upload process is implemented as a series of requests to upload the file in chunks. The library
provides a single method that wraps the steps required to upload your file. Please refer to the
[API Reference](https://docs.nightfall.ai/reference) for more details.

The file is uploaded synchronously, but as files can be arbitrarily large, the scan itself is conducted asynchronously.
The results from the scan are delivered by webhook; for more information about setting up a webhook server, refer to
[the docs](https://docs.nightfall.ai/docs/creating-a-webhook-server).

#### Sample Code

See [examples/file/file\_scanner.go](examples/file/file_scanner.go) for an example

