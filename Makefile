build:
	docker build -t tail-based-sampling:latest .
	docker tag tail-based-sampling:latest registry.cn-shanghai.aliyuncs.com/mingowong/tail-based-sampling:latest
	docker push registry.cn-shanghai.aliyuncs.com/mingowong/tail-based-sampling:latest
