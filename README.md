# Key-Value Storage Engine

## About this repository

This repository contains the source code that I wrote for the
[Build Your Own Key-Value Storage Engine](https://read.thecoder.cafe/p/build-your-own-kv-engine)
series of articles published to [The Coder Cafe](https://read.thecoder.cafe/) on
[Substack](https://substack.com). The article series, written by
[Teive Harsanyi](https://substack.com/@teivah) is intended to be an 8-week
challenge to teach participants about the skills necessary to build a modern
key-value storage engine.

This repository contains the source code for my solution. I am sharing the
source code publicly on GitHub to help others who may be taking this challenge
or who are interested in doing so, but might be a little afraid of taking on
the challenge if they have holes in their understanding of computer science and
data structures and algorithms. Hopefully, studying my code will help you
through the tougher parts of the challenge and help you to better understand the
concepts and grow professionally.

## Blog posts

I am writing about the implementation of my solution to the challenge on my
[blog](https://michaelfcollins3.dev). Here's a list of the published articles
that cover the implementation to date:

* [Building a Key-Value Storage Engine: Week 1](building-a-key-value-storage-engine-week-1)

## Getting started

### Prerequisites

* [Git](https://git-scm.com)
* [GitHub CLI](https://cli.github.com)
* [Go 1.27.1](https://go.dev)

### Cloning the repository

To work with the source code and run the key-value storage engine locally, you
should clone the GitHub repository onto your computer. To do this, open a
terminal and navigate to the location in your file system where you want to host
the repository. Run:

```shell
gh repo clone mfcollins3/kvstore
cd kvstore
```

This command will clone the repository into the `kvstore` subdirectory and will
check out the `main` branch.

### Running the Key-Value Storage Engine

To run the key-value storage engine locally, run the following command in the
terminal:

```shell
go run ./cmd/server
```

This command will start the server and the server will listen for incoming
requests at https://localhost:8080.

### Testing the Key-Value Storage Engine

To test the key-value storage engine, you will first need to create test data.
This can be done using the [generator program](cmd/gen/main.go). In a terminal,
run:

```shell
go run ./cmd/gen put 30000
```

This will create the `put.txt` file with 30,000 test operations. Next, with the
server running, run the client program to process `put.txt`:

```shell
go run ./cmd/client
```

The client program will read `put.txt`, process the operations and call the
server, and will output any error results.