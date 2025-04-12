



#描述：pvm 是什么？

pvm 是一个 PHP 版本管理器，用于在 Windows 系统上管理多个 PHP 版本。

###主要功能：
列出已安装的 PHP 版本：
pvm list - 显示所有已安装的 PHP 版本，包括已映射和未映射的版本。
安装指定版本的 PHP：
pvm install <版本> - 从 Windows PHP 官方网站下载并安装指定版本的 PHP。例如：pvm install 7.4。
切换 PHP 版本：
pvm use <版本> - 切换系统使用的 PHP 版本，这将更新系统的 PATH 环境变量。例如：pvm use 7.4。
显示帮助信息：
pvm - 不带参数时显示帮助信息。
#pvm 的工作原理：
它将所有 PHP 版本安装在 D:\app\pvm\phps 目录中
通过修改系统 PATH 环境变量来切换 PHP 版本
自动从 Windows PHP 官方仓库下载适合 Windows 的 PHP 版本
支持解压缩 PHP 压缩包并自动配置基本设置
除了以上基本命令外，程序还有一些额外功能：
根据指定的版本号自动查找最新的匹配版本
自动创建和配置 php.ini 文件
为当前会话和系统级别更新 PATH 环境变量
你目前可以使用 pvm list 查看已安装的版本，使用 pvm install <版本> 安装新版本，以及 pvm use <版本> 切换到特定版本。
