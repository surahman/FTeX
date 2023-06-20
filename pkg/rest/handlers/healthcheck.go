package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/surahman/FTeX/pkg/common"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/redis"
)

// Healthcheck checks if the service is healthy by pinging the data tier comprised of Postgres and Redis.
//
//	@Summary		Healthcheck for service liveness.
//	@Description	This endpoint is exposed to allow load balancers etc. to check the health of the service.
//	@Description	This is achieved by the service pinging the data tier comprised of Postgres and Redis.
//	@Tags			health healthcheck liveness
//	@Id				healthcheck
//	@Produce		json
//	@Success		200	{object}	models.HTTPSuccess	"message: healthy"
//	@Failure		503	{object}	models.HTTPError	"error message with any available details"
//	@Router			/health [get]
func Healthcheck(logger *logger.Logger, db postgres.Postgres, cache redis.Redis) gin.HandlerFunc {
	return func(context *gin.Context) {
		httpStatus, httpMsg, _ := common.HTTPHealthcheck(db, cache, logger)

		context.JSON(httpStatus, &models.HTTPSuccess{Message: httpMsg})
	}
}
