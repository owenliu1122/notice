package cmd

import (
	"fmt"
	"message_notification_practice"
	"message_notification_practice/controllers"
	"message_notification_practice/mq"
	"message_notification_practice/pb"
	"message_notification_practice/services"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	log "gopkg.in/cihub/seelog.v2"
)

var serverCmdPort int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start up message notification server",
	Run:   serverProc,
}

func serverProc(cmd *cobra.Command, args []string) {

	log.Debugf("redis：%#v\n", viper.GetStringMap("server"))

	var cfg notice.Config

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Printf("%s service Unmarshal configuration is failed, err: %s", cmd.Use, err.Error())
		return
	}

	log.Debug("Start serverProc")
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	db, err := gorm.Open("mysql", cfg.Server.MySQL)
	if err != nil {
		fmt.Printf("init mysql failed, err: %s", err)
		return
	}

	defer db.Close()

	// grpc server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverCmdPort))
	handleInitError(err, "net")

	gs := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10 * time.Minute,
		}),
	)

	defer gs.GracefulStop()

	mqConnection, err := mq.NewConnection(cfg.Server.RabbitMQ)
	if err != nil {
		log.Error("new rabbitmq connection failed, err: ", err)
		return
	}
	defer mqConnection.Close()

	producer, err := mq.NewProducer("server producer", mqConnection)
	if err != nil {
		log.Error("NewProducer failed, err: ", err)
		return
	}
	defer producer.Close()

	svc := services.NewNotificationService(db,
		producer,
		cfg.Server.Producer.Exchange,
		cfg.Server.Producer.RoutingKey)

	ctl := controllers.NewServerController(svc)
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
