package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/skinkvi/effective_mobile/internal/storage/postgres"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	storage *postgres.Storage
	logger  *zap.Logger
}

func New(storage *postgres.Storage, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		storage: storage,
		logger:  logger,
	}
}

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" binding:"required" example:"Yandex Plus"`
	Price       int       `json:"price" binding:"required" example:"400"`
	UserID      uuid.UUID `json:"user_id" binding:"required" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" binding:"required" example:"07-2025"`
	EndDate     *string   `json:"end_date" example:"08-2025"`
}

// / @Summary Create a new subscription
//
//	@Description	Create a new subscription
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			subscription	body		CreateSubscriptionRequest	true	"Subscription to create"
//	@Success		201				{object}	map[string]int				"id"
//	@Failure		400				{object}	map[string]string			"error"
//	@Failure		500				{object}	map[string]string			"error"
//	@Router			/subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var subReq CreateSubscriptionRequest

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&subReq); err != nil {
		h.logger.Error("failed to bind JSON", zap.Error(err), zap.String("request_body", string(bodyBytes)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("CreateSubscription request", zap.String("service_name", subReq.ServiceName), zap.Int("price", subReq.Price), zap.Any("user_id", subReq.UserID), zap.String("start_date", subReq.StartDate), zap.Any("end_date", subReq.EndDate))

	startDate, err := time.Parse("01-2006", subReq.StartDate)
	if err != nil {
		h.logger.Error("failed to parse start date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	var endDate *time.Time
	if subReq.EndDate != nil {
		endDateVal, err := time.Parse("01-2006", *subReq.EndDate)
		if err != nil {
			h.logger.Error("failed to parse end date", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
			return
		}
		endDate = &endDateVal
	}

	sub := postgres.Subscription{
		ServiceName: subReq.ServiceName,
		Price:       subReq.Price,
		UserID:      subReq.UserID,
		StartDate:   &startDate,
		EndDate:     endDate,
	}

	id, err := h.storage.CreateSubscription(c.Request.Context(), sub)
	if err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// / @Summary Get a subscription by ID
//
//	@Description	Get a subscription by ID
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path		int	true	"Subscription ID"
//	@Success		200	{object}	postgres.Subscription
//	@Failure		400	{object}	map[string]string	"error"
//	@Failure		404	{object}	map[string]string	"error"
//	@Router			/subscriptions/{id} [get]

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idParam := c.Param("id")
	h.logger.Info("GetSubscription request", zap.String("id", idParam))
	if idParam == "" {
		h.logger.Error("subscription ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		h.logger.Error("invalid subscription ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	sub, err := h.storage.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	response := gin.H{
		"id":           sub.ID,
		"service_name": sub.ServiceName,
		"price":        sub.Price,
		"user_id":      sub.UserID,
	}

	if sub.StartDate != nil {
		response["start_date"] = sub.StartDate.Format("01-2006")
	}

	if sub.EndDate != nil {
		response["end_date"] = sub.EndDate.Format("01-2006")
	}

	c.JSON(http.StatusOK, response)
}

// / @Summary Update a subscription
//	@Description	Update a subscription
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int						true	"Subscription ID"
//	@Param			subscription	body		postgres.Subscription	true	"Subscription to update"
//	@Success		200				{object}	map[string]string		"message"
//	@Failure		400				{object}	map[string]string		"error"
//	@Failure		404				{object}	map[string]string		"error"
//	@Failure		500				{object}	map[string]string		"error"
//	@Router			/subscriptions/{id} [put]

type UpdateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int       `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     *string   `json:"end_date" example:"08-2025"`
}

// / @Summary Update a subscription
//
//	@Description	Update a subscription
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int							true	"Subscription ID"
//	@Param			subscription	body		UpdateSubscriptionRequest	true	"Subscription to update"
//	@Success		200				{object}	map[string]string			"message"
//	@Failure		400				{object}	map[string]string			"error"
//	@Failure		404				{object}	map[string]string			"error"
//	@Failure		500				{object}	map[string]string			"error"
//	@Router			/subscriptions/{id} [put]

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idParam := c.Param("id")
	h.logger.Info("UpdateSubscription request", zap.String("id", idParam))
	if idParam == "" {
		h.logger.Error("subscription ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		h.logger.Error("invalid subscription ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	var subReq UpdateSubscriptionRequest

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&subReq); err != nil {
		h.logger.Error("failed to bind JSON", zap.Error(err), zap.String("request_body", string(bodyBytes)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("UpdateSubscription request body", zap.Int("id", id), zap.String("service_name", subReq.ServiceName), zap.Int("price", subReq.Price), zap.Any("user_id", subReq.UserID), zap.String("start_date", subReq.StartDate), zap.Any("end_date", subReq.EndDate))

	sub, err := h.storage.GetSubscription(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription for update", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if subReq.ServiceName != "" {
		sub.ServiceName = subReq.ServiceName
	}

	if subReq.Price != 0 {
		sub.Price = subReq.Price
	}

	if subReq.UserID != (uuid.UUID{}) {
		sub.UserID = subReq.UserID
	}

	if subReq.StartDate != "" {
		startDate, err := time.Parse("01-2006", subReq.StartDate)
		if err != nil {
			h.logger.Error("failed to parse start date", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
			return
		}
		sub.StartDate = &startDate
	}

	if subReq.EndDate != nil {
		endDate, err := time.Parse("01-2006", *subReq.EndDate)
		if err != nil {
			h.logger.Error("failed to parse end date", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
			return
		}
		sub.EndDate = &endDate
	}

	if err := h.storage.UpdateSubscription(c.Request.Context(), sub); err != nil {
		h.logger.Error("failed to update subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated successfully"})
}

// / @Summary Delete a subscription
//
//	@Description	Delete a subscription
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path		int					true	"Subscription ID"
//	@Success		200	{object}	map[string]string	"message"
//	@Failure		400	{object}	map[string]string	"error"
//	@Failure		404	{object}	map[string]string	"error"
//	@Router			/subscriptions/{id} [delete]

func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idParam := c.Param("id")
	h.logger.Info("DeleteSubscription request", zap.String("id", idParam))
	if idParam == "" {
		h.logger.Error("subscription ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		h.logger.Error("invalid subscription ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.storage.DeleteSubscription(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete subscription", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted successfully"})
}

// / @Summary List subscriptions
//
//	@Description	List subscriptions
//	@Tags			subscriptions
//	@Produce		json
//	@Param			user_id			query		string	false	"User ID"
//	@Param			service_name	query		string	false	"Service Name"
//	@Success		200				{array}		postgres.Subscription
//	@Failure		500				{object}	map[string]string	"error"
//	@Router			/subscriptions [get]

func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")
	h.logger.Info("ListSubscriptions request", zap.String("user_id", userIDStr), zap.String("service_name", serviceName))

	var userID *uuid.UUID
	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Error("invalid user ID", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
		userID = &parsedUserID
	}

	var serviceNamePtr *string
	if serviceName != "" {
		serviceNamePtr = &serviceName
	}

	subs, err := h.storage.ListSubscriptions(c.Request.Context(), userID, serviceNamePtr)
	if err != nil {
		h.logger.Error("failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	var response []gin.H
	for _, sub := range subs {
		subResp := gin.H{
			"id":           sub.ID,
			"service_name": sub.ServiceName,
			"price":        sub.Price,
			"user_id":      sub.UserID,
		}

		if sub.StartDate != nil {
			subResp["start_date"] = sub.StartDate.Format("01-2006")
		}

		if sub.EndDate != nil {
			subResp["end_date"] = sub.EndDate.Format("01-2006")
		}

		response = append(response, subResp)
	}

	c.JSON(http.StatusOK, response)
}

// / @Summary Calculate total cost of subscriptions
//
//	@Description	Calculate total cost of subscriptions
//	@Tags			subscriptions
//	@Produce		json
//	@Param			user_id			query		string				true	"User ID"
//	@Param			service_name	query		string				true	"Service Name"
//	@Param			start_date		query		string				true	"Start Date"
//	@Param			end_date		query		string				true	"End Date"
//	@Success		200				{object}	map[string]int		"total_cost"
//	@Failure		400				{object}	map[string]string	"error"
//	@Failure		500				{object}	map[string]string	"error"
//	@Router			/subscriptions/total_cost [get]

func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	h.logger.Info("CalculateTotalCost request", zap.String("user_id", userIDStr), zap.String("service_name", serviceName), zap.String("start_date", startDateStr), zap.String("end_date", endDateStr))

	if userIDStr == "" || serviceName == "" || startDateStr == "" || endDateStr == "" {
		h.logger.Error("user ID, service name, start date, and end date are required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID, service name, start date, and end date are required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format. user ID must be a valid UUID"})
		return
	}

	startDate, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		h.logger.Error("invalid start date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	endDate, err := time.Parse("01-2006", endDateStr)
	if err != nil {
		h.logger.Error("invalid end date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	totalCost, err := h.storage.CalculateTotalCost(c.Request.Context(), startDate, endDate, userID, serviceName)
	if err != nil {
		h.logger.Error("failed to calculate total cost", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate total cost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_cost": totalCost})
}
