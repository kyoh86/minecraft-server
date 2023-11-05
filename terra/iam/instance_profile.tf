// Minecraft Java Serverインスタンス用のInstance Profileを作成
data "aws_iam_policy_document" "ec2_assume_role" {
  statement {
    actions = [
      "sts:AssumeRole"
    ]
    principals {
      type = "Service"
      identifiers = [
        "ec2.amazonaws.com"
      ]
    }
  }
}

resource "aws_iam_role" "instance_role" {
  name               = "instance_role"
  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
  managed_policy_arns = [
    // - SSM接続を許可する
    "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  ]
}

# Role attachment: SSM接続の許可
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
  name   = "instance_runnable"
  policy = data.aws_iam_policy_document.instance_runnable.json
}

# Role attachment: RunInstances
resource "aws_iam_role_policy_attachment" "instance_runnable" {
  role       = aws_iam_role.instance_role.name
  policy_arn = aws_iam_policy.instance_runnable.arn
}

# Instance Profile
resource "aws_iam_instance_profile" "instance_profile" {
  name = "instance_profile"
  role = aws_iam_role.instance_role.name
}
