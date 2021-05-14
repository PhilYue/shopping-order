/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/logger"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	"github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/protocol/dubbo"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/nacos"
	"github.com/transaction-wg/seata-golang/pkg/client"
	seataConfig "github.com/transaction-wg/seata-golang/pkg/client/config"
	"github.com/transaction-wg/seata-golang/pkg/client/tm"
)

import (
	"github.com/apache/dubbo-go-samples/shopping-order/go-client/pkg"
)

const (
	SEATA_CONF_FILE = "SEATA_CONF_FILE"
	CONF_CONSUMER_FILE_PATH = "CONF_CONSUMER_FILE_PATH"
	APP_LOG_CONF_FILE = "APP_LOG_CONF_FILE"
)

var (
	survivalTimeout int = 10e9
	debuging bool = true
	step, command string
	sleeps int
)

// they are necessary:
// 		export CONF_CONSUMER_FILE_PATH="xxx"
// 		export APP_LOG_CONF_FILE="xxx"
//      export SEATA_CONF_FILE="xxx"

func main() {
	// seata-golang init
	confFile := os.Getenv(SEATA_CONF_FILE)
	seataConfig.InitConf(confFile)
	client.NewRpcClient()
	tm.Implement(pkg.ProxySvc)

	// dubbo-go init
	config.Load()

	time.Sleep(3e9)

	fmt.Print(">>> Enable debug (default `true`): ")
	fmt.Scanln(&debuging)

	if debuging {
		debug()
	} else {

		// commit success
		pkg.ProxySvc.CreateSo(context.TODO(), false)

		// rollback process
		pkg.ProxySvc.CreateSo(context.TODO(), true)

		initSignal()
	}
}

func debug() {
	for {
		fmt.Println(">>>Begin the funny practice useing dubbogo and seata-golang!")
		fmt.Print(">>>Select mode `normal`-(distributed transaction commit success) or `exception`-(distributed transaction rollback):")
		fmt.Scanln(&step)
		if step == "normal" {
			fmt.Println(">>>Current mode `normal`-(distributed transaction commit success)...")

			// commit success
			pkg.ProxySvc.CreateSoDebug(context.TODO(), false, true)

			fmt.Println(">>>The distributed transaction has been committed!")
		} else if step == "exception" {
			fmt.Println(">>>Current mode `exception`-(distributed transaction rollback)...")

			// rollback process
			pkg.ProxySvc.CreateSoDebug(context.TODO(), true, true)

			fmt.Println(">>>The distributed transaction has been rolled back!")
		} else {
			fmt.Println(">>>Unknown command, maybe you want try the `normal` pattern or the `exception` pattern!\n\n\n")
		}
		step = ""
	}
}

func initSignal() {
	signals := make(chan os.Signal, 1)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP,
		syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-signals
		logger.Infof("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				logger.Warnf("app exit now by force...")
				os.Exit(1)
			})

			// The program exits normally or timeout forcibly exits.
			fmt.Println("app exit now...")
			return
		}
	}
}
