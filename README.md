# easyeda-git (lceda-git) 嘉立创 EDA 专业版 版本控制/团队协作工具

**This is a third-party open source tool**

**这是一个第三方开源工具**

easyeda-git (lceda-git) is another collaborate tool for EasyEDA Pro. It can help
engineer teams to convert the EasyEDA eprj binary file to human readable SQL
code. Expert engineer teams will be able to easily diff the changes, and all the
change history can be stored in a version control system (e.g. git).

easyeda-git (lceda-git) 是一个嘉立创 EDA 专业版的版本控制和团队协作工具，它可以
帮助工程师团队将嘉立创 EDA 专业版的 eprj 二进制工程文件转化为人类可读的 SQL 代码
，以便专业的工程团队检查工程的更改，及使用代码管理软件（如 git）进行版本控制

## Getting Started (EasyEDA plugin users) 开始使用（嘉立创EDA专业版插件）

Not yet available 暂不支持

## Getting Started (GUI users) 开始使用（图形化界面）

Not yet available 暂不支持

## Getting Started (CLI users) 开始使用（命令行）

### Installation Method 1: Download the prebuilt CLI tool 安装方法一：下载命令行程序

Not yet available 暂不支持

### Installation Method 2: Install with golang module (compile from source) 安装方法二：使用 Go 安装（即源码编译安装）

```bash
$ go install github.com/neuroflag/easyeda-git
```

Don't forget to add `$HOME/go/bin` in your system PATH

不要忘记将 `$HOME/go/bin`添加至系统的 PATH 环境变量

### Usage 使用方法

All your need is saving your project locally, closing EasyEDA Pro and to run

```bash
$ easyeda-git path/to/MyProject.eprj
```

只需要在本地保存eprj工程，退出后执行

```bash
$ lceda-git path/to/MyProject.eprj
```

You will see the code is extracted

你可以看到本工具首先从工程中提取出了代码

```bash
$ git status -s

A path/to/MyProject.eprj.sql
A path/to/MyProject_Board1_SCH1.eprj.sql
A path/to/MyProject_Board1_SCH2.eprj.sql
A path/to/MyProject_Board1_PCB.eprj.sql
A path/to/MyProject_SCH1.eprj.sql
A path/to/MyProject_PCB.eprj.sql
```

Then project is constructed again from SQL files

然后本工具从代码恢复了工程

Then project is opened with EasyEDA Pro again

然后本工具启动了嘉立创EDA专业版，打开了恢复的工程

For further usage, please run `easyeda-git --help`

其他使用方法，请运行 `lceda-git --help`查看

## Made with love by Neuroflag

EasyEDA by JLC is an extrodinary tool that simplifies the PCB design. Engineers
don't need to worry about where to find the chip symbols or where to buy the
chips. All the engineers need is to focus on designing the PCB. However, the
version control feature is extremely confusing for our devops team. The missing
feature (like Altium's ascii schematics) makes the tool inpropable for large
team and projects. After a little bit investiating, we found that the EasyEDA
eprj file is a simple sqlite3 database, which makes version control with SQL
possible. So the answer is simple. Make the tool, share it with the EasyEDA
community, and wait for this need is officially supported.
