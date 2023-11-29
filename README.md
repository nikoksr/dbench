<div align="center">

&nbsp;
<h1>dbench</h1>
<p><i>A nifty wrapper around pgbench that comes with plotting and result management.</i></p>

&nbsp;

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/nikoksr/dbench)
</div>

&nbsp;

## About

`dbench` is a convenient wrapper around `pgbench` that enhances your benchmarking experience with features like result
management and plotting. It's designed to make it easy to run, manage, and visualize your PostgreSQL benchmarks.

## Installation

> Important: While the releases offer binaries for multiple platforms and architectures, only Linux is tested. If you
> are using a different OS, I do not guarantee that dbench will work as expected.

Download one of the [releases](https://github.com/nikoksr/dbench/releases) for your system, or install using the
provided [install script](scripts/install.sh):

```sh
curl -fsSL https://tinyurl.com/install-dbench | bash
```

## Prerequisites

`dbench` requires `pgbench` and `gnuplot` to be installed on your system. You can check if they are installed and their
versions using the `dbench doctor` command.

## Usage

> Note: To enhance security, dbench does not offer a password flag. Instead, you have two options: either set the
> PGPASSWORD environment variable, or input your password when prompted. dbench will subsequently use the PGPASSWORD
> environment variable in its sub-processes.

To use `dbench`, you first need to initialize a PostgreSQL Database. Remember to adjust the connection parameters to
your needs.

```bash
dbench init --db-name=postgres --db-user=postgres --db-host=localhost --db-port=5432 --scale 10
```

Then, you can run your first array of benchmarks.

```bash
dbench run --db-name=postgres --db-user=postgres --db-host=localhost --db-port=5432
```

Afterward, you can plot the results.

```bash
dbench plot <id>
```

To see all available commands and flags, run `dbench --help`.
