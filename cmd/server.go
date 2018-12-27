package cmd

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	log "gopkg.in/cihub/seelog.v2"
	"message_notification_practice/controllers"
	"message_notification_practice/mq"
	"message_notification_practice/pb"
	"message_notification_practice/services"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var serverCmdPort int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start up message notification server",
	Run:   serverProc,
}

func serverProc(cmd *cobra.Command, args []string) {

	log.Debug("Start serverProc")
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	db, err := gorm.Open("mysql", "root:123456@/msg_notification?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Printf("init mysql failed, err: %s", err)
		return
	}

	defer db.Close()

	time.Sleep(2 * time.Second) // TODO: remove is

	// grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverCmdPort))
	handleInitError(err, "net")

	gs := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10 * time.Minute,
		}),

		//// Register stream middleware.
		//grpc.StreamInterceptor(controllers.ClientIDSetter),
	)

	defer gs.GracefulStop()

	pcCfg := mq.MQCfg{
		URL:      "amqp://liujx:Liujiaxing@localhost:5672/",
		Queue:    "push.msg.q",
		Exchange: "t.msg.ex",
	}

	mqCli := mq.NewMq(pcCfg)
	if e := mqCli.InitConnection(); e != nil {
		log.Error("InitConnection failed, err: ", e)
	}
	defer mqCli.Close()

	if e := mqCli.InitProducer(pcCfg.Exchange, pcCfg.Queue); e != nil {
		log.Error("InitProducer failed, err: ", e)
	}

	ctl := controllers.NewServerController(services.NewNotificationService(db, mqCli))
	pb.RegisterMsgNotificationServer(gs, ctl)
	go gs.Serve(lis)

	log.Info("server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Debug("Exit serverProc")
}

func init() {

	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().IntVarP(&serverCmdPort, "port", "p", 3000, "Port to listen")

}
