@echo off
REM คัดลอกไฟล์นี้เป็น run.bat แล้วแก้ค่าด้านล่าง (run.bat ถูกกันไว้ใน .gitignore จะไม่ขึ้น GitHub)
cd /d %~dp0

REM คีย์ลับสำหรับเข้ารหัส session ยาวอย่างน้อย 32 ตัว สุ่มด้วยคำสั่ง PowerShell:
REM   [guid]::NewGuid().ToString("N") + [guid]::NewGuid().ToString("N")
set SESSION_KEY=ใส่คีย์ที่สุ่มเองตรงนี้

REM การเชื่อมต่อฐานข้อมูล (ผู้ใช้:รหัสผ่าน@tcp(โฮสต์:พอร์ต)/ชื่อฐาน)
set DB_DSN=root:@tcp(127.0.0.1:3306)/pitch_booking?charset=utf8mb4

REM เปิดสองบรรทัดนี้เฉพาะตอนเปิดผ่าน HTTPS (เช่น Cloudflare Tunnel)
REM set COOKIE_SECURE=1
REM set TRUST_PROXY=1

go run .
pause
