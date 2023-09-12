package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	grbook "booksearch/bkfind"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr   = flag.String("addr", "localhost:8082", "the address to connect to")
	author = flag.String("author", "", "Author's name")
	book   = flag.String("book", "", "The title of the book")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer conn.Close()

	client := grbook.NewFinderClient(conn)

	if *author != "" && *book != "" {
		log.Fatal("Cannot use both flags \"author\" and \"book\"")
	}

	if *author == "" && *book == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if *author != "" {
		resp, err := client.FindBooks(ctx, &grbook.BReq{Author: *author})
		if err != nil {
			log.Fatalf("could not acquire: %v", err)
		}
		fmt.Printf("%s", resp.GetBooks())
	}

	if *book != "" {
		resp, err := client.FindAuthors(ctx, &grbook.AReq{Book: *book})
		if err != nil {
			log.Fatalf("could not acquire: %v", err)
		}
		fmt.Printf("%s", resp.GetAuthors())
	}
}
