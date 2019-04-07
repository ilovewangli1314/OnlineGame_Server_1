package main

import (
	"flag"
	"fmt"

	"strings"

	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya"
	"github.com/topfreegames/pitaya/acceptor"
	"github.com/topfreegames/pitaya/component"
	"github.com/topfreegames/pitaya/serialize/protobuf"
	"github.com/ilovewangli1314/OnlineGame_Server_1/services"
)

func configureBackend() {
	room := services.NewRoom()
	pitaya.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)

	pitaya.RegisterRemote(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)
}

func configureFrontend(port int) {
	ws := acceptor.NewWSAcceptor(fmt.Sprintf(":%d", port))
	pitaya.Register(&services.Connector{},
		component.WithName("connector"),
		component.WithNameFunc(strings.ToLower),
	)
	pitaya.RegisterRemote(&services.ConnectorRemote{},
		component.WithName("connectorremote"),
		component.WithNameFunc(strings.ToLower),
	)

	pitaya.AddAcceptor(ws)
}

func main() {
	defer pitaya.Shutdown()

	port := flag.Int("port", 3250, "the port to listen")
	svType := flag.String("type", "connector", "the server type")
	isFrontend := flag.Bool("frontend", true, "if server is frontend")
	flag.Parse()

	ser := protobuf.NewSerializer()
	pitaya.SetSerializer(ser)

	// if !*isFrontend {
	// 	configureBackend()
	// } else {
	// 	configureFrontend(*port)
	// }
	room := services.NewRoom()
	pitaya.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)
	ws := acceptor.NewWSAcceptor(fmt.Sprintf(":%d", port))
	pitaya.AddAcceptor(ws)
	// pitaya.RegisterRemote(room,
	// 	component.WithName("room"),
	// 	component.WithNameFunc(strings.ToLower),
	// )

	conf := configApp()
	pitaya.Configure(*isFrontend, *svType, pitaya.Cluster, map[string]string{}, conf)
	pitaya.Start()
}

func configApp() *viper.Viper {
	conf := viper.New()
	// conf.SetEnvPrefix("chat") // allows using env vars in the CHAT_PITAYA_ format
	conf.SetDefault("pitaya.buffer.handler.localprocess", 15)
	conf.Set("pitaya.heartbeat.interval", "15s")
	conf.Set("pitaya.buffer.agent.messages", 32)
	conf.Set("pitaya.handler.messages.compression", false)
	return conf
}
