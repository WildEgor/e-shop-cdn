package handlers

import (
	delete_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/delete"
	download_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/download"
	error_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/errors"
	get_files_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/get_files"
	health_check_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/health_check"
	metadata_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/metadata"
	ready_check_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ready_check"
	replace_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/replace"
	ticker_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ticker"
	upload_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/upload"
	ws_connect_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ws_connect"
	ws_disconnect_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ws_disconnect"
	handshake_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ws_handshake"
	sub_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ws_subscribe"
	unsub_handler "github.com/WildEgor/e-shop-cdn/internal/handlers/ws_unsubscribe"
	"github.com/WildEgor/e-shop-cdn/internal/repositories"
	"github.com/google/wire"
)

// HandlersSet contains http/amqp/etc handlers (acts like facades)
var HandlersSet = wire.NewSet(
	repositories.RepositoriesSet,

	error_handler.NewErrorsHandler,
	health_check_handler.NewHealthCheckHandler,
	ready_check_handler.NewReadyCheckHandler,
	upload_handler.NewUploadHandler,
	download_handler.NewDownloadHandler,
	delete_handler.NewDeleteHandler,
	metadata_handler.NewMetadataHandler,
	get_files_handler.NewGetFilesHandler,
	replace_handler.NewReplaceHandler,
	sub_handler.NewSubscribeHandler,
	unsub_handler.NewUnsubscribeHandler,
	handshake_handler.NewWSHandshakeHandler,
	ws_connect_handler.NewWSConnectHandler,
	ws_disconnect_handler.NewWSDisconnectHandler,

	// HINT: for testing only
	ticker_handler.NewTickerHandler,
)
