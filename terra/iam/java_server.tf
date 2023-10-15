// Minecraft Java Serverインスタンス用のInstance Profileを作成
// - SSM接続を許可する
resource "aws_iam_role" "java_server_role" {
  name               = "java_server_role"
  assume_role_policy = file("./assume_role_policy.json")
}

# Role attachment: SSM接続の許可
resource "aws_iam_role_policy_attachment" "java_server_attach" {
  role       = aws_iam_role.java_server_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

data "aws_iam_policy_document" "instance_runnable" {
  statement {
    sid = "1"
    actions = [
      "ec2:RunInstances"
    ]
    resources = [
      "*"
    ]
  }
}

# Policy: RunInstances
resource "aws_iam_policy" "instance_runnable" {
  policy = data.aws_iam_policy_document.instance_runnable.json
}

# Role attachment: RunInstances
resource "aws_iam_role_policy_attachment" "instance_runnable" {
  role       = aws_iam_role.java_server_role.name
  policy_arn = aws_iam_policy.instance_runnable.arn
}

# Instance Profile
resource "aws_iam_instance_profile" "java_server_profile" {
  name = "java_server_profile"
  role = aws_iam_role.java_server_role.name
}
