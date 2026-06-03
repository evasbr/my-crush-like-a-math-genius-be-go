package controller

import (
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ProductController struct {
	service.ProductService
	configuration.Config
	Redis *redis.Client
	log   *logrus.Entry
}

func NewProductController(productService *service.ProductService, config configuration.Config, redis *redis.Client) *ProductController {
	return &ProductController{
		ProductService: *productService,
		Config:         config,
		Redis:          redis,
		log:            common.Log.WithField("scope", "ProductController"),
	}
}

func (controller ProductController) Route(router fiber.Router) {
	product := router.Group("/product")
	product.Post("/", middleware.RequireAuth([]string{"ROLE_ADMIN", "write:product"}, controller.Config, controller.Redis), controller.Create)
	product.Put("/:id", middleware.RequireAuth([]string{"ROLE_ADMIN", "write:product"}, controller.Config, controller.Redis), controller.Update)
	product.Delete("/:id", middleware.RequireAuth([]string{"ROLE_ADMIN", "write:product"}, controller.Config, controller.Redis), controller.Delete)
	product.Get("/:id", middleware.RequireAuth([]string{"ROLE_ADMIN", "read:product"}, controller.Config, controller.Redis), controller.FindById)
	product.Get("/", middleware.RequireAuth([]string{"ROLE_ADMIN", "read:product"}, controller.Config, controller.Redis), controller.FindAll)
}

// Create func create product.
// @Description create product.
// @Summary create product
// @Tags Product
// @Accept json
// @Produce json
// @Param request body model.ProductCreateOrUpdateModel true "Request Body"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/product [post]
func (controller ProductController) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	controller.log.WithContext(ctx).Info("Menerima request POST /v1/api/product")

	var request model.ProductCreateOrUpdateModel
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	response := controller.ProductService.Create(ctx, request)
	return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
		Code:    fiber.StatusCreated,
		Message: "Success",
		Data:    response,
	})
}

// Update func update one exists product.
// @Description update one exists product.
// @Summary update one exists product
// @Tags Product
// @Accept json
// @Produce json
// @Param request body model.ProductCreateOrUpdateModel true "Request Body"
// @Param id path string true "Product Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/product/{id} [put]
func (controller ProductController) Update(c *fiber.Ctx) error {
	var request model.ProductCreateOrUpdateModel
	id := c.Params("id")
	err := c.BodyParser(&request)
	exception.PanicLogging(err)

	response := controller.ProductService.Update(c.UserContext(), request, id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    response,
	})
}

// Delete func delete one exists product.
// @Description delete one exists product.
// @Summary delete one exists product
// @Tags Product
// @Accept json
// @Produce json
// @Param id path string true "Product Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/product/{id} [delete]
func (controller ProductController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	controller.ProductService.Delete(c.UserContext(), id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
	})
}

// FindById func gets one exists product.
// @Description Get one exists product.
// @Summary get one exists product
// @Tags Product
// @Accept json
// @Produce json
// @Param id path string true "Product Id"
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/product/{id} [get]
func (controller ProductController) FindById(c *fiber.Ctx) error {
	id := c.Params("id")

	result := controller.ProductService.FindById(c.UserContext(), id)
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}

// FindAll func gets all exists products.
// @Description Get all exists products.
// @Summary get all exists products
// @Tags Product
// @Accept json
// @Produce json
// @Success 200 {object} model.GeneralResponse
// @Security JWT
// @Router /v1/api/product [get]
func (controller ProductController) FindAll(c *fiber.Ctx) error {
	result := controller.ProductService.FindAll(c.UserContext())
	return c.Status(fiber.StatusOK).JSON(model.GeneralResponse{
		Code:    200,
		Message: "Success",
		Data:    result,
	})
}
