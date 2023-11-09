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

### 基本的なツールの準備

```
sudo su - ec2-user

sudo yum update -y
sudo yum install -y xfsprogs udev java-17-amazon-corretto
sudo yum clean all

sudo amazon-linux-extras install -y nginx1
sudo cp -a /etc/nginx/nginx.conf /etc/nginx/nginx.conf.back
sudo systemctl enable nginx
```

### EBSの初期化

ファイルシステム作るやつ。当然データ消えるので注意。

```
sudo su - ec2-user

sudo mkfs -t xfs /dev/sdh
```

### Minecraft実行用ユーザーの準備

```
sudo su - ec2-user

sudo adduser minecraft --gid wheel
```

### Minecraft格納用領域のマウント

Minecraft実行用ユーザーの所有として領域を作る

```
sudo su - ec2-user

sudo mkdir -p /minecraft
sudo mount /dev/sdh /minecraft
sudo chown minecraft:wheel /minecraft
```

### Minecraft本体、Modのインストール

```
sudo su - minecraft

cd /minecraft

MINECRAFT_VERSION=1.20.2

# Install "FabricMC" as a Minecraft Server
FABRIC_LOADER_VERSION=0.14.24
FABRIC_INSTALLER_VERSION=0.11.1

curl -Lo ./fabric-server-mc.jar https://meta.fabricmc.net/v2/versions/loader/${MINECRAFT_VERSION}/${FABRIC_LOADER_VERSION}/${FABRIC_INSTALLER_VERSION}/server/jar
java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting
mkdir -p mods

# Install "Fabric API" to run Fabric Mods
FABRIC_API_VERSION=0.90.7
curl -Lo ./mods/fabric-api.jar https://github.com/FabricMC/fabric/releases/download/${FABRIC_API_VERSION}+${MINECRAFT_VERSION}/fabric-api-${FABRIC_API_VERSION}+${MINECRAFT_VERSION}.jar

# Install "GeyserMC" to cross-play between Bedrock & Java
curl -Lo ./mods/Geyser-Fabric.jar https://download.geysermc.org/v2/projects/geyser/versions/latest/builds/latest/downloads/fabric

# Install "DiscordIntegration" to integrate with Discord
DISCORD_INTEGRATION_VERSION=3.0.3
curl -Lo ./mods/dcintegration-fabric.jar https://cdn.modrinth.com/data/rbJ7eS5V/versions/ZlLJC9ox/dcintegration-fabric-${DISCORD_INTEGRATION_VERSION}-${MINECRAFT_VERSION}.jar

# Install "Floodgate" to support login with Bedrock account
curl -Lo ./mods/floodgate-fabric.jar https://ci.opencollab.dev/job/GeyserMC/job/Floodgate-Fabric/job/master/lastSuccessfulBuild/artifact/build/libs/floodgate-fabric.jar

# Install "Dynmap" to show world map
curl -Lo ./mods/Dynmap-fabric.jar https://www.curseforge.com/api/v1/mods/59433/files/4765921/download

# Initialize Fabric server settings
timeout 1m java -Xmx2G -jar fabric-server-mc.jar --nogui --initSetting || :
```

### 設定系

以下ファイルをどうにかして置く。まあVimで開いてコピペで良いと思う。
SSHとかはリスク無駄にでかいのでやらない。

- nginx.conf → /etc/nginx/nginx.conf
- docker/data/ 下の色々 → /minecraft下のファイルに一つ一つ上書きしていく

## 動かす

screenで充分。
detachは`<C-a>+d`
