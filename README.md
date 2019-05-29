
# gophercoin

A simple cryptocurrency and blockchain implementation

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for various purposererequisites

In order to be able to install and run the application you will need the following programs in your machine.

```
go version >= go1.12.1

npm
```

### Installing

You can install the app by running the following command in the root folder of the project.

```
make install
```

### Running

Considering you installed the app by running the previous step

```
#The following command will launch the app regularly
gcd

#The following command will launch the app with the REST API
# on port 9050
gcd -rest 9050

# The following command will launch the app with the REST API
# on port 9050 and with the blockchain.db file in the current directory as database
gcd -rest 9050 -db blockchain.db

```





### Testing

Writing the command `gcd -h` will print out a more detailed explanation of how the input flags work.
```
Usage of gcd:
  -addr string
    	Address used for mining reward.
  -db string
    	Path to the blockchain.db file.
  -listen string
    	Port for the daemon to use to listen for peer connections
  -mining true
    	Set to true to mine, `false` not to.
  -rest string
    	Port to use for the REST API server.
  -wallet string
    	Path to the wallet.dat file.

```
Considering you have an instance with the REST API running

```
#The following command will create a wallet
curl -H "Content-Type: application/json" -X POST http://127.0.0.1:9050/create_wallet     

#The following command will create the blockchain
curl -H "Content-Type: application/json" -X POST http://127.0.0.1:9050/create_blockchain    

```

## API Endpoints


The following API endpoints are exposed when the daemon is executed with the REST flag
```
	
    "GET",
    "/",

    "GET",
    "/new_address",
	
    "POST",
    "/create_wallet",

    "POST",
    "/create_blockchain",

    "GET",
    "/get_balance/{Address}",

    "GET",
    "/list_addresses",
	
    "GET",
    "/list_mempool",
	
    "GET",
    "/list_blocks",
	
    "GET",
    "/node_info",
	
    "POST",
    "/submit_tx/{From}/{To}/{Amount}",
	
    "POST",
    "/add_node/{Address}",

```
## Built With

* [golang](https://golang.org) - The programming language


