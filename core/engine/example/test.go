package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func example2() {
	mux := &MyMux{}
	http.ListenAndServe(":80", mux)
}
