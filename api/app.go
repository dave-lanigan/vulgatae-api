package app

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/dave-lanigan/vulgatae-api/api/database"
)

func Run() {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(cors.New())

	initDB()
	defer database.DBConn.Close()

	app.Get("/", func(c *fiber.Ctx) error {

		metaData := VulgataeMeta{
			AboutAPI: "API for the Vulgate/Douay-Rheims Bible",
			About:    "The Vulgate is a version of the Bible",
			Contact:  "bohemdev@tutanota.com",
		}

		return c.JSON(metaData)
	})

	app.Get("/books", func(c *fiber.Ctx) error {
		db := database.DBConn
		var books []Book
		db.Find(&books)
		return c.JSON(books)
	})

	app.Get("/books/:book", func(c *fiber.Ctx) error {
		tag := c.Params("book")
		db := database.DBConn
		var book Book
		db.Where("books.tag=?", tag).Find(&book)
		return c.JSON(book)
	})

	app.Get("/books/:book/chapters", func(c *fiber.Ctx) error {
		book := c.Params("book")
		type result struct {
			Number int    `json:"number"`
			Long   string `json:"bookLong"`
			Header string `json:"header"`
		}

		res := []result{}

		db := database.DBConn
		db.Model(&Chapter{}).
			Select("chapters.number, books.long, chapters.header").
			Where("books.tag=?", book).
			Joins("LEFT JOIN books ON books.bid = chapters.book").
			Scan(&res)
		return c.JSON(res)
	})

	app.Get("/books/:book/:chapter", func(c *fiber.Ctx) error {
		return c.SendString("Chapter content.")
	})

	app.Get("/books/:book/:chapter/:verse", func(c *fiber.Ctx) error {

		book := c.Params("book")
		chapter := c.Params("chapter")
		verse := c.Params("verse")

		type result struct {
			Long    string `json:"bookLong"`
			Number  int    `json:"chapter"`
			Latin   string `json:"latin"`
			English string `json:"english"`
		}

		res := []result{}

		db := database.DBConn

		db.Model(&Verse{}).
			Select("books.long, chapters.number, verses.latin, verses.english").
			Where("books.tag=?", book).
			Where("verses.chapter=?", chapter).
			Where("verses.number=?", verse).
			Joins("LEFT JOIN books ON books.bid = verses.book").
			Joins("LEFT JOIN chapters ON chapters.cid = verses.chapter").
			Scan(&res)
		return c.JSON(res)
	})

	log.Fatal(app.Listen(":3000"))
}

func initDB() {
	var err error
	database.DBConn, err = gorm.Open("sqlite3", "v.db")

	if err != nil {
		panic("Failed to connect to DB!")
	}
}

type VulgataeMeta struct {
	AboutAPI string `json:"aboutAPI"`
	About    string `json:"about"`
	Contact  string `json:"contact"`
}

type Edition struct {
	Eid  int    `gorm:"primary_key"`
	Name string `json:"name"`
	Date string `json:"date"`
	Info string `json:"info"`
}

type Book struct {
	Bid       int    `json:"bid"`
	Number    int    `json:"number"`
	Short     string `json:"short"`
	Long      string `json:"long"`
	Alt       string `json:"alt"`
	Tag       string `json:"tag"`
	Blurb     string `json:"blurb"`
	Testament string `json:"testament"`
	Edition   string `json:"edition"`
}

type Chapter struct {
	Number int    `json:"number"`
	Book   int    `json:"book"`
	Header string `json:"header"`
}

type Verse struct {
	Number     int    `json:"number"`
	Book       int    `json:"book"`
	Chapter    int    `json:"chapter"`
	Latin      string `json:"latin"`
	English    string `json:"english"`
	Commentary string `json:"commentary"`
}
