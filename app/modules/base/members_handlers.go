package base

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type inviteMemberRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	DisplayName string   `json:"display_name" binding:"required"`
	Roles       []string `json:"roles"`
}

type updateRolesRequest struct {
	Roles []string `json:"roles" binding:"required,min=1"`
}

type updateMemberStatusRequest struct {
	Status MembershipStatus `json:"status" binding:"required"`
}

func (s *AuthService) listMembersHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	members, err := s.ListMembers(c.Request.Context(), principal.TenantID)
	if err != nil {
		slog.Error("list members failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list members"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": members})
}

func (s *AuthService) listRolesHandler(c *gin.Context) {
	roles, err := s.ListRoles(c.Request.Context())
	if err != nil {
		slog.Error("list roles failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (s *AuthService) inviteMemberHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request inviteMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and display name are required"})
		return
	}

	member, err := s.InviteMember(c.Request.Context(), principal, InviteMemberInput{
		Email: request.Email, DisplayName: request.DisplayName, RoleCodes: request.Roles,
	})
	if err != nil {
		respondMemberError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": member})
}

func (s *AuthService) updateMemberRolesHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	membershipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid membership id"})
		return
	}

	var request updateRolesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one role is required"})
		return
	}

	member, err := s.UpdateMemberRoles(c.Request.Context(), principal, membershipID, request.Roles)
	if err != nil {
		respondMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": member})
}

func (s *AuthService) updateMemberStatusHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	membershipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid membership id"})
		return
	}

	var request updateMemberStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	member, err := s.SetMemberStatus(c.Request.Context(), principal, membershipID, request.Status)
	if err != nil {
		respondMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": member})
}

func respondMemberError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
	case errors.Is(err, ErrAlreadyMember):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrLastOwner), errors.Is(err, ErrCannotSelfAlter):
		// Both are deliberate refusals to let an administrator lock the tenant
		// out of itself, not validation failures.
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUnknownRole), errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		slog.Error("membership request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "membership request failed"})
	}
}
