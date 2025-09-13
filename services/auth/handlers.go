package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LoginByPhone(c *gin.Context) {
	var req PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	loginResponse, err := h.userService.LoginByPhone(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		switch err.Error() {
		case "phone not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Phone not found"})
		case "invalid credentials":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case "phone must start with international calling code prefix '+'":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone must start with international calling code prefix '+'"})
		case "invalid phone number format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		case "password must be at least 8 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		case "password must be at most 32 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at most 32 characters long"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

func (h *UserHandler) RegisterByPhone(c *gin.Context) {
	fmt.Println("Handler: RegisterByPhone")
	var req PhoneAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.userService.RegisterByPhone(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		switch err.Error() {
		case "phone number already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already exists"})
		case "phone must start with international calling code prefix '+'":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone must start with international calling code prefix '+'"})
		case "invalid phone number format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		case "password must be at least 8 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		case "password must be at most 32 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at most 32 characters long"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) LoginWithEmail(c *gin.Context) {
	fmt.Println("Handler: LoginWithEmail")
	var req EmailAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	loginResponse, err := h.userService.LoginWithEmail(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err.Error() {
		case "email not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		case "invalid credentials":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case "invalid email address format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
		case "password must be at least 8 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		case "password must be at most 32 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at most 32 characters long"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, loginResponse)
}

func (h *UserHandler) RegisterWithEmail(c *gin.Context) {
	fmt.Println("Handler: RegisterWithEmail")
	var req EmailAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.userService.RegisterWithEmail(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err.Error() {
		case "email address already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "Email address already exists"})
		case "invalid email address format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
		case "password must be at least 8 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		case "password must be at most 32 characters long":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at most 32 characters long"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) LinkPhone(c *gin.Context) {
	var req LinkPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else if len(authHeader) > 6 && authHeader[:6] == "Barer " {
		tokenString = authHeader[6:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	claims, err := h.userService.jwtService.ValidateToken(tokenString)
	if err != nil {
		switch err.Error() {
		case "token has expired":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
		case "invalid token":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		}
		return
	}

	resp, err := h.userService.LinkPhone(c.Request.Context(), claims.UserID, req.Phone)
	if err != nil {
		switch err.Error() {
		case "phone is taken":
			c.JSON(http.StatusConflict, gin.H{"error": "Phone is taken"})
		case "phone must start with international calling code prefix '+'":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Phone must start with international calling code prefix '+'"})
		case "invalid phone number format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone number format"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
