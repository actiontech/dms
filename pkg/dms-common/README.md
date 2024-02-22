# dms-common

1. 校验插件接口：./api/dms/v1/plugin.go
2. 反向代理接口：./api/dms/v1/proxy.go
3. 用户查询接口：./api/dms/v1/user.go
4. 用户查询工具：./dmsobject/user.go
5. 校验插件&反向代理注册工具：./register/register.go
6. 事前权限校验涉及的数据结构，背景见 https://github.com/actiontech/dms-ee/issues/125：./sql_op 