package main

import (
	"log"
	"net/http"
	Controller "prevent-race-condition/controller"
	"prevent-race-condition/domain"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	//Create DB Connection
	dsn := "root:dev@tcp(127.0.0.1:3304)/booking-system?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("error connection :", err)
	}

	//Create Migration
	db.Migrator().DropTable(&domain.Order{}, &domain.Stock{})
	db.AutoMigrate(&domain.Order{}, &domain.Stock{})

	//Create Seeder
	if err := db.Create(&domain.Stock{FlightId: "CGK001-0293", Stock: 10, Class: "Economy"}).Error; err != nil {
		log.Println("error Create Seeder:", err)
	}

	// Define Controller
	controller := Controller.NewOrderController(db)

	//Router
	r := mux.NewRouter()
	r.HandleFunc("/race-condition", controller.CreateRaceCondition).Methods(http.MethodPost)
	r.HandleFunc("/no-race-condition", controller.CreateNoRaceCondition).Methods(http.MethodPost)

	log.Println("Application run on port 8081")
	http.ListenAndServe(":8081", r)
}
