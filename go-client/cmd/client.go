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
	_ "github.com/apache/dubbo-go/registry/zookeeper"
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
	debuging bool = false
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

	fmt.Println(">>>是否开启 debug ：")
	fmt.Scanln(&debuging)

	if debuging {
		debug()
	} else {

		// commit success
		//pkg.ProxySvc.CreateSo(context.TODO(), false)

		// rollback process
		pkg.ProxySvc.CreateSo(context.TODO(), true)

		initSignal()
	}



	// debug



}

func debug() {
	for {
		fmt.Println(">>>Begin the funny practice useing dubbogo and seata-golang!")
		fmt.Print(">>>选择实战模式 `normal`-正常事务提交 or `exception`-异常事务回滚 ：")
		fmt.Scanln(&step)
		if step == "normal" {
			fmt.Println(">>>当前模式 `normal`-正常事务提交")

			// commit success
			pkg.ProxySvc.CreateSoDebug(context.TODO(), false, true)

			fmt.Println(">>>the distributed transaction has been committed!")
		} else if step == "exception" {
			fmt.Println(">>>当前模式 `exception`-异常事务回滚")

			// rollback process
			pkg.ProxySvc.CreateSoDebug(context.TODO(), true, true)

			fmt.Println(">>>simulation debug, let's see what happens...")
			time.Sleep(time.Second * 10)
			//select {}
			fmt.Println(">>>simulation debug, input command `next` ")
			// process next
			fmt.Scanln(&command)
			if command == "next" {
				// commit or rollback
			}

			fmt.Println(">>>ok , maybe the distributed transaction has been rolled back, let's see what happens again...")
			time.Sleep(time.Second * 10)
			fmt.Println(">>>the distributed transaction has been rolled back!")
		} else {
			fmt.Println(">>>unknown command, maybe you want try the `normal` pattern or the `exception` pattern!\n\n\n")
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
