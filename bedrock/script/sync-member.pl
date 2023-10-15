#!/usr/bin/perl

use strict;
use warnings;
use utf8;
use open ":utf8";

binmode STDIN, ":utf8";
binmode STDOUT, ":utf8";

my @data = qx"curl --silent --location https://docs.google.com/spreadsheets/d/1g9bhugOcgFIbVLuCWg3C8IcWp9-BaHvzAve3l7yUoJg/export?format=csv";
my @allowlist;
my @permissions;

foreach (@data) {
	my @cells = split(/,/);
	if ($#cells < 3) {
		next;
	}
	my $name = $cells[0];
	my $xuid = $cells[1];
	my $perm = $cells[2];
	$name =~ s/^\s+|\s+$//g;
	$xuid =~ s/^\s+|\s+$//g;
	if ($name eq '' || $xuid eq '' || $name eq 'ID' || $xuid eq 'XUID') {
		next;
	}
	push(@allowlist, sprintf('{"name": "%s", "xuid": "%d"}', $name, $xuid));
	push(@permissions, sprintf('{"xuid": "%s", "permission": "%s"}', $xuid, $perm));
}

my $allow_out;
open($allow_out, ">", "/opt/minecraft_be_server/allowlist.json") || die("Failed to open allowlist.json");
printf $allow_out "[%s]", join(",", @allowlist);
close($allow_out);

my $perm_out;
open($perm_out, ">", "/opt/minecraft_be_server/permissions.json") || die("Failed to open permissions.json");
printf $perm_out "[%s]", join(",", @permissions);
close($perm_out);
