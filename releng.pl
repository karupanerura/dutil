use strict;
use warnings;

use Getopt::Long qw/:config posix_default no_ignore_case bundling/;
use File::Temp 0.15 qw/tempfile/;
use Term::ReadLine;

my ($major, $minor, $patch);
GetOptions('major' => \$major, 'minor' => \$minor, 'patch' => \$patch)
    or die "Usage: $0 [--major] [--minor] [--patch]";
die "Usage: $0 [--major] [--minor] [--patch]" unless $major || $minor || $patch;

sub confirm {
    my $message = shift;

    my $term = Term::ReadLine->new('confirm');
    while (defined(my $line = $term->readline($message.' [y/n]: '))) {
        return !!1 if $line eq 'y';
        return !!0 if $line eq 'n';
    }
}

sub parse_version {
    my $file = shift;

    open my $fh, '<', $file
        or die "$!: $file";

    my %version;
    my $in_const;
    while (defined(my $line = <$fh>)) {
        chomp $line;
        if ($in_const) {
            next if $line eq '';
            if ($line =~ /^\)$/) {
                $in_const = !!0;
            } elsif ($line =~ /^\s*(\w+)\s*=\s*(\d+)\s*$/) {
                $version{lc $1} = 0+$2;
            } else {
                die "unknonw line: $line";
            }
        } else {
            next if $line =~ /^package/;
            next if $line eq '';
            if ($line =~ /^const/) {
                $in_const = !!1;
            } else {
                die "unknonw line: $line";
            }
        }
    }
    return \%version;
}

sub rewrite_version {
    my ($file, $version) = @_;

    open my $fh, '<', $file
        or die "$!: $file";
    my ($wfh, $tempfile) = tempfile(UNLINK => 1);

    my $in_const;
    while (defined(my $line = <$fh>)) {
        chomp $line;
        if ($in_const) {
            if ($line eq '') {
                print $wfh $/;
            } elsif ($line =~ /^\)$/) {
                print $wfh $line, $/;
                $in_const = !!0;
            } elsif ($line =~ /^(\s*)(\w+)(\s*)=(\s*)\d+(\s*)$/) {
                print $wfh $1, $2, $3, '=', $4, $version->{lc $2}, $5, $/;
            } else {
                die "unknonw line: $line";
            }
        } else {
            print $wfh $line, $/;
            next if $line =~ /^package/;
            next if $line eq '';
            if ($line =~ /^const/) {
                $in_const = !!1;
            } else {
                die "unknonw line: $line";
            }
        }
    }
    rename $tempfile, $file
        or die $!;
}

sub format_version {
    my $version = shift;
    sprintf '%d.%d.%d', @{$version}{qw/major minor patch/};
}

my $v = parse_version('./internal/version/number.go');
my $curr = format_version($v);
$v->{major}++ if $major;
$v->{minor}++ if $minor;
$v->{patch}++ if $patch;
my $next = format_version($v);
print "$curr -> $next", $/;
die 'aborted' unless confirm('OK?');

rewrite_version('./internal/version/number.go', $v);
system $^X, '-i', '-pe', "s/$curr/$next/g", 'README.md';
system 'git', 'add', '.';
system 'git', 'diff', '--cached';
die 'aborted' unless confirm('OK?');

system 'git', 'commit', '-m', "v$next";
system 'git', 'tag', "v$next";