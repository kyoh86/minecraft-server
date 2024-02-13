# Minecraft server

## AWSの準備

AWS内はTerraformで構成している。

- `terra/volume` EBSのストレージを作っている。データの永続化に関わるので、消すとデータ全部消えるので注意。
- `terra/instance` EC2のインスタンスを建てている。インストールなんかの手順は吹き飛ぶので、これも注意。
- `terra/iam` IAMのロール類。

## インスタンスへの接続

./find-instance.zsh でInstance IDがわかるので、`aws ssm start-session --target "Instance ID"`でつなぐ

## インスタンスの整備

インスタンス内は手でセットアップしている。
以下作業は`terra/instance`内で、インスタンスIDを変数にぶち込んで作業している。

```console
$ INSTANCE_ID="$(terraform output -json | jq -r '.instance.value')"
$ echo $INSTANCE_ID
```

### EBSの初期化

ファイルシステム作るやつ。**当然データ消えるので注意。EBS初めて作った時にしかやらない**。

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo yum update -y
$ sudo yum install -y xfsprogs udev
$ sudo yum clean all
$ sudo mkfs -t xfs /dev/sdh
```

### 基本的なツールの準備

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo yum update -y
$ sudo yum install -y java-17-amazon-corretto
$ sudo yum clean all

$ sudo amazon-linux-extras install -y nginx1
```

### Minecraft格納用領域の準備

Minecraft実行用ユーザーの所有として領域を作る

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo adduser minecraft --gid wheel

$ sudo mkdir -p /minecraft
$ sudo mount /dev/sdh /minecraft
```

EBSのデータが消えてる場合は、以下の通りchownをかける。

```console
$ sudo chown minecraft:wheel /minecraft
```

### Minecraft本体、Modのインストール

EBSのデータが残ってる限りはいらない。

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo su - minecraft

$ cd /minecraft

$ MINECRAFT_VERSION=1.20.4

$ : # Install "FabricMC" as a Minecraft Server
$ FABRIC_LOADER_VERSION=0.15.6
$ FABRIC_INSTALLER_VERSION=1.0.0

$ curl -Lo ./fabric-server-mc.jar https://meta.fabricmc.net/v2/versions/loader/${MINECRAFT_VERSION}/${FABRIC_LOADER_VERSION}/${FABRIC_INSTALLER_VERSION}/server/jar
$ java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting
$ mkdir -p mods

$ : # Install "Fabric API" to run Fabric Mods
$ FABRIC_API_VERSION=0.96.0
$ curl -Lo ./mods/fabric-api.jar https://github.com/FabricMC/fabric/releases/download/${FABRIC_API_VERSION}+${MINECRAFT_VERSION}/fabric-api-${FABRIC_API_VERSION}+${MINECRAFT_VERSION}.jar

$ : # Install "GeyserMC" to cross-play between Bedrock & Java
$ curl -Lo ./mods/Geyser-Fabric.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/fabric

$ : # Install "DiscordIntegration" to integrate with Discord
$ curl -Lo ./mods/dcintegration-fabric.jar https://cdn.modrinth.com/data/rbJ7eS5V/versions/t7t8BbD7/dcintegration-fabric-3.0.5-1.20.4.jar

$ : # Install "Floodgate" to support login with Bedrock account
$ curl -Lo ./mods/floodgate-fabric.jar https://ci.opencollab.dev/job/GeyserMC/job/Floodgate-Fabric/job/master/lastSuccessfulBuild/artifact/build/libs/floodgate-fabric.jar

$ : # Install "Dynmap" to show world map
$ curl -Lo ./mods/Dynmap-fabric.jar https://cdn.modrinth.com/data/fRQREgAc/versions/fOe507oy/Dynmap-3.7-beta-4-fabric-1.20.4.jar

$ : # Initialize Fabric server settings
$ timeout 1m java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting || :
```

### 設定系

以下ファイルを順次置いていく。

#### アップロード

```console
$ scp -r -i ~/.ssh/minecraft_instance ../../data "${INSTANCE_ID}:/home/ec2-user/data"
```

#### 設定

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo rm -rf /home/minecraft/data
$ sudo mv /home/ec2-user/data /home/minecraft/data
$ sudo chown -R minecraft:wheel /home/minecraft/data
$ sudo cp -a /etc/nginx/nginx.conf /etc/nginx/nginx.conf.bak
$ cat /home/minecraft/data/nginx/nginx.conf | sudo tee /etc/nginx/nginx.conf
$ cat /home/minecraft/data/systemd/minecraft.service | sudo tee /etc/systemd/system/minecraft.service
```

Minecraftの設定系は、さらにminecraftユーザーに切り替えて、逐次更新していく

```console
$ sudo su - minecraft
$ ls ~/data
$ ls /minecraft

... 順次上書きするなり転記するなりする
```

## 動かす

`ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"` でつないでやる

```console
$ sudo systemctl daemon-reload
$ sudo systemctl enable nginx
$ sudo systemctl start nginx
$ sudo systemctl enable minecraft
$ sudo systemctl start minecraft
```

## インスタンスへの接続

```console
$ aws sso login
$ export INSTANCE_ID="$(terraform -chdir=./terra/instance/ output -json | jq -r '.instance.value')"
$ ssh -i ~/.ssh/minecraft_instance "${INSTANCE_ID}"
```
