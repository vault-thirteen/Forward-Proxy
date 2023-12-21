# Forward Proxy

Forward proxy server for _HTTP_ and _HTTPS_ protocols. 

## Supported Features
* Forward-proxying _HTTP_ data streams.
* Forward-proxying _HTTPS_ data streams.
* Ability to unpack _Gzipped_ data streams.
* Ability to detect and remove _Unicode_ BOM (Byte Order Mark).
* Ability to limit the speed for both _HTTP_ and _HTTPS_ data streams.
* Configurable listen host name and port number.
* Two work modes: public & private.
* White list of IP addresses is supported.
* Usage of interfaces implementing `io.Reader` interface.
* Pure Golang solution, free and open-source.

## Building
Use the `build.bat` script included with the source code.

## Installation
`go install github.com/vault-thirteen/Forward-Proxy/cmd/proxy@latest`  

## Startup Parameters
| Parameter |  Type   | Description                                   | Possible Values                                        |     Unit     | Default Value |
|:---------:|:-------:|-----------------------------------------------|--------------------------------------------------------|:------------:|:-------------:|
|   -bom    | Boolean | Remove BOM from content                       |                                                        |              |     true      |
|   -gzip   | Boolean | Decode GZip content                           |                                                        |              |     false     |
|   -host   | String  | Listen host name                              |                                                        |              |   "0.0.0.0"   |
|   -list   | String  | Path to a list of IP addresses                |                                                        |              |      ""       |
| -loglevel | String  | Log level                                     | debug, info, warn, error, fatal, panic, none, disabled |              |    "error"    |
|   -mode   | String  | Work mode                                     | public, private                                        |              |   "public"    |
|   -port   | Integer | Listen port number                            |                                                        |              |     8080      |
|    -sl    | Boolean | Use speed limiter                             |                                                        |              |     true      |
|   -slbl   | Integer | Speed limiter's burst limit                   |                                                        | bytes / sec. |    50'000     |
|  -slbnr   |  Float  | Speed limiter's maximal burst-to-normal ratio |                                                        |              |      2.0      |
|   -slnl   |  Float  | Speed limiter's normal limit                  |                                                        | bytes / sec. |    50'000     |
|   -tcdt   | Integer | Target connection dial timeout                |                                                        |     sec.     |      60       |

### Notes
* To get help, use `-h` startup parameter. 


* List of IP addresses has different usage depending on the work mode:
  * In public mode, list is not used at all;
  * In private mode, list is used as a white list of IP addresses.


* Limiting speed to values lower than 32 KiB/sec. (32'768 Bytes/sec.) is not 
supported due to restrictions of the `io.Copy` function built into _Go_ language.
This limit may change in future versions of _Golang_.
