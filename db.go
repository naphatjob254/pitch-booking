package main

// db.go – ทุกอย่างที่เกี่ยวกับฐานข้อมูล: การเชื่อมต่อ, สร้างตารางอัตโนมัติ, โมเดลข้อมูล และคำสั่งค้นหาที่ใช้ร่วมกัน

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

// ==========================================
// โมเดลข้อมูล (Structs)
// ==========================================
type MatchJoiner struct {
	Username string
	Phone    string // จะมีค่าเฉพาะตอนผู้ดูเป็นเจ้าของโพสต์
}

type Match struct {
	ID            int
	UserID        int
	OwnerName     string
	OwnerPhone    string
	TeamName      string
	Title         string
	PitchType     string
	PlayersNeeded int
	CurrentJoin   int
	HasJoined     bool
	IsOwner       bool
	Joiners       []MatchJoiner
}

type Pitch struct {
	ID           int
	Name         string
	PitchType    string
	PricePerHour int
	AvgRating    float64
	ReviewCount  int
	Stars        string
}

type Review struct {
	ID        int
	Username  string
	Rating    int
	Comment   string
	Stars     string
	UpdatedAt string
	IsOwner   bool
}

type AdminReview struct {
	ID        int
	Username  string
	PitchName string
	Rating    int
	Stars     string
	Comment   string
}

type Booking struct {
	ID          int
	PitchName   string
	BookingDate string
	StartTime   string
	Status      string
}

type AdminMatch struct {
	ID            int
	Username      string
	TeamName      string
	Title         string
	PitchType     string
	PlayersNeeded int
}

type AdminBooking struct {
	ID          int
	Username    string
	Phone       string
	PitchName   string
	BookingDate string
	StartTime   string
	Status      string
}

// ==========================================
// เชื่อมต่อฐานข้อมูล + เตรียมตาราง
// ==========================================
func initDB() {
	var err error
	dsn := env("DB_DSN", "root:@tcp(127.0.0.1:3306)/pitch_booking?charset=utf8mb4")
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("DB Connect Error: ", err)
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatal("DB Ping Error: ", err)
	}
	ensureSchema()
	seedPitches()
}

// คำสั่งสร้างตาราง (IF NOT EXISTS: ถ้ามีตารางอยู่แล้วจะไม่แตะต้องข้อมูลเดิม)
var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		username   VARCHAR(30)  NOT NULL UNIQUE,
		phone      VARCHAR(15)  NOT NULL,
		password   VARCHAR(100) NOT NULL,
		role       ENUM('user','admin') NOT NULL DEFAULT 'user',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

	`CREATE TABLE IF NOT EXISTS pitches (
		id             INT AUTO_INCREMENT PRIMARY KEY,
		name           VARCHAR(50) NOT NULL,
		pitch_type     VARCHAR(30) NOT NULL,
		price_per_hour INT NOT NULL,
		image_url      VARCHAR(255) NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

	`CREATE TABLE IF NOT EXISTS matches (
		id             INT AUTO_INCREMENT PRIMARY KEY,
		user_id        INT NOT NULL,
		team_name      VARCHAR(50)  NOT NULL,
		title          VARCHAR(100) NOT NULL,
		pitch_type     VARCHAR(100) NOT NULL,
		players_needed TINYINT UNSIGNED NOT NULL,
		created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

	`CREATE TABLE IF NOT EXISTS match_joins (
		match_id INT NOT NULL,
		user_id  INT NOT NULL,
		PRIMARY KEY (match_id, user_id),
		FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id)  REFERENCES users(id)   ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

	`CREATE TABLE IF NOT EXISTS bookings (
		id            INT AUTO_INCREMENT PRIMARY KEY,
		user_id       INT NOT NULL,
		pitch_id      INT NOT NULL,
		booking_date  DATE NOT NULL,
		start_time    TIME NOT NULL,
		contact_phone VARCHAR(15) NOT NULL,
		status        VARCHAR(20) NOT NULL DEFAULT 'Pending',
		UNIQUE KEY uq_slot (pitch_id, booking_date, start_time),
		FOREIGN KEY (user_id)  REFERENCES users(id)   ON DELETE CASCADE,
		FOREIGN KEY (pitch_id) REFERENCES pitches(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

	`CREATE TABLE IF NOT EXISTS reviews (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		pitch_id   INT NOT NULL,
		user_id    INT NOT NULL,
		rating     TINYINT UNSIGNED NOT NULL,
		comment    VARCHAR(500) NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uq_review (pitch_id, user_id),
		CHECK (rating BETWEEN 1 AND 5),
		FOREIGN KEY (pitch_id) REFERENCES pitches(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id)  REFERENCES users(id)   ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
}

func ensureSchema() {
	for _, q := range schemaStatements {
		if _, err := db.Exec(q); err != nil {
			log.Fatal("สร้างตารางไม่สำเร็จ: ", err)
		}
	}
}

// ถ้ายังไม่มีสนามเลย ใส่สนามตัวอย่าง 3 สนาม (แก้ราคา/ชื่อทีหลังได้ใน phpMyAdmin)
func seedPitches() {
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM pitches").Scan(&n); err != nil {
		log.Println("นับสนามไม่สำเร็จ:", err)
		return
	}
	if n > 0 {
		return
	}
	samples := []struct {
		name  string
		kind  string
		price int
	}{
		{"สนาม A", "7 คน", 500},
		{"สนาม B", "9 คน", 700},
		{"สนาม C", "11 คน", 1000},
	}
	for _, s := range samples {
		if _, err := db.Exec("INSERT INTO pitches (name, pitch_type, price_per_hour) VALUES (?, ?, ?)", s.name, s.kind, s.price); err != nil {
			log.Println("เพิ่มสนามตัวอย่างไม่สำเร็จ:", err)
		}
	}
	log.Println("เพิ่มสนามตัวอย่าง 3 สนามแล้ว")
}

// ==========================================
// ตัวช่วยเกี่ยวกับ error / คำสั่งที่ใช้ซ้ำ
// ==========================================

// isDup: true เมื่อ MySQL ตอบว่าข้อมูลซ้ำ (UNIQUE) รหัส 1062
func isDup(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

// loadPitches: ดึงสนามทั้งหมดพร้อมคะแนนรีวิวเฉลี่ย
func loadPitches() ([]Pitch, error) {
	rows, err := db.Query(`SELECT p.id, p.name, p.pitch_type, p.price_per_hour,
			COALESCE(ROUND(AVG(r.rating), 1), 0), COUNT(r.id)
		FROM pitches p LEFT JOIN reviews r ON r.pitch_id = p.id
		GROUP BY p.id, p.name, p.pitch_type, p.price_per_hour ORDER BY p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pitches := []Pitch{}
	for rows.Next() {
		var p Pitch
		if err := rows.Scan(&p.ID, &p.Name, &p.PitchType, &p.PricePerHour, &p.AvgRating, &p.ReviewCount); err != nil {
			return nil, err
		}
		p.Stars = stars(int(math.Round(p.AvgRating)))
		pitches = append(pitches, p)
	}
	return pitches, rows.Err()
}
