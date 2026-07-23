package logger

import (
	"go.uber.org/zap"
)

// глобальный логгер
var Log *zap.SugaredLogger

func init() { // функция запускается автоматически при ипорте пакета
	// создаём предустановленный регистратор zap
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer zapLogger.Sync()
	// делаем регистратор SugaredLogger
	Log = zapLogger.Sugar()
}
