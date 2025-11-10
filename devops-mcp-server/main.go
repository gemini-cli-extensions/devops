// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	httpAddr  = flag.String("http", "", "if set, use streamable HTTP at this address, instead of stdin/stdout. e.g. localhost:8080")
	pprofAddr = flag.String("pprof", "", "if set, host the pprof debugging server at this address")
)

func main() {
	flag.Parse()

	if *pprofAddr != "" {
		// For debugging memory leaks, add an endpoint to trigger a few garbage
		// collection cycles and ensure the /debug/pprof/heap endpoint only reports
		// reachable memory.
		http.DefaultServeMux.Handle("/gc", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			for range 3 {
				runtime.GC()
			}
			fmt.Fprintln(w, "GC'ed")
		}))
		go func() {
			// DefaultServeMux was mutated by the /debug/pprof import.
			if err := http.ListenAndServe(*pprofAddr, http.DefaultServeMux); err != nil {
				log.Printf("pprof server failed: %v", err)
			}
		}()
	}

	server := createServer()

	// Serve over stdio, or streamable HTTP if -http is set.
	if *httpAddr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		log.Printf("MCP handler listening at %s", *httpAddr)
		if *pprofAddr != "" {
			log.Printf("pprof listening at http://%s/debug/pprof", *pprofAddr)
		}
		log.Fatal(http.ListenAndServe(*httpAddr, handler))
	} else {
		//Default server is stdio.
		t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
		if err := server.Run(context.Background(), t); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}
}
