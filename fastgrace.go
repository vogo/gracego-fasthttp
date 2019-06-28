package main

import (
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/vogo/gracego"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

const (
	serviceName = "fastgrace"
	addr        = ":8082"
)

var (
	server *fasthttp.Server
)

func main() {
	agentRouter := router.New()
	agentRouter.GET("/ok", OKHandler)
	agentRouter.GET("/shutdown", ShutdownHandler)
	agentRouter.GET("/upgrade", UpgradeHandler)
	agentRouter.GET("/download.zip", DownloadHandler)

	server = &fasthttp.Server{
		Handler:            agentRouter.Handler,
		ReadTimeout:        3600 * time.Second,
		WriteTimeout:       3600 * time.Second,
		MaxRequestBodySize: 1 << 20,
	}


	go func() {
		t := time.NewTicker(2 * time.Second)
		s := reflect.ValueOf(server)
		f := s.Elem().FieldByName("open")

		for {
			<-t.C

			info("fasthttp open: %d", f.Int())
		}
	}()

	info("start %s at %s", serviceName, addr)
	err := gracego.Serve(server, serviceName, addr)
	info("server end. %+v", err)
}

func ShutdownHandler(ctx *fasthttp.RequestCtx) {
	info("request shutdown")
	go func() {
		info("do shutdown")
		err := server.Shutdown()
		info("shutdown finish. %+v", err)
	}()
}

func OKHandler(ctx *fasthttp.RequestCtx) {
	info("request ok")
	_, _ = ctx.WriteString("ok")
}

func info(format string, args ...interface{}) {
	log.Println(gracego.GetServerID(), "-", fmt.Sprintf(format, args...))
}

//DownloadHandler download the graceup server zip
func DownloadHandler(ctx *fasthttp.RequestCtx) {
	info("request download")
	path, err := os.Executable()
	if err != nil {
		responseError(ctx, err)
		return
	}

	dir := filepath.Dir(path)
	zipFilePath := fmt.Sprintf("%s%c%s", dir, os.PathSeparator, serviceName+".zip")
	file, err := os.OpenFile(zipFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		responseError(ctx, err)
		return
	}

	ctx.Response.Header.Add("content-type", "application/octet-stream")
	_, err = io.Copy(ctx, file)
	if err != nil {
		responseError(ctx, err)
		return
	}
}

//UpgradeHandler restart server
func UpgradeHandler(ctx *fasthttp.RequestCtx) {
	info("request upgrade")
	err := gracego.Upgrade("v2", serviceName, "http://127.0.0.1"+addr+"/download.zip")
	if err != nil {
		responseError(ctx, err)
	}

	_, _ = ctx.WriteString("success")
}

func responseError(ctx *fasthttp.RequestCtx, e error) {
	ctx.Response.Header.Add("content-type", "plain/text")
	ctx.Response.SetStatusCode(400)
	_, _ = ctx.WriteString(e.Error())
}
