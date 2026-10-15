package main

import "html/template"

// loadTemplates รวมหน้าเว็บทั้งหมด (เก็บเป็นค่าคงที่ในไฟล์ ui_*.go) ให้เป็นชุดเทมเพลตเดียว
// ชื่อที่ใช้เรียกใน render(...) ต้องตรงกับชื่อ key ด้านล่าง
func loadTemplates() *template.Template {
	pages := map[string]string{
		"partials.html": partialsPage,
		"login.html": loginPage,
		"register.html": registerPage,
		"dashboard.html": dashboardPage,
		"form.html": formPage,
		"pitches.html": pitchesPage,
		"book.html": bookPage,
		"my_bookings.html": myBookingsPage,
		"reviews.html": reviewsPage,
		"admin.html": adminPage,
	}
	t := template.New("root")
	for name, src := range pages {
		template.Must(t.New(name).Parse(src))
	}
	return t
}
