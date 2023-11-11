<div align="center">

&nbsp;
<h1>dbench</h1>
<p><i>A convenience wrapper around pgbench that adds benchmarks persistence and plotting.</i></p>

&nbsp;

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/nikoksr/dbench)
</div>

&nbsp;
## About <a id="about"></a>

> Warning: At the time of writing, this project should be considered a PoC or barely usable alpha. It is not recommended to take the results of this benchmark too seriously.

DBench is a convenience wrapper around [pgbench](https://www.postgresql.org/docs/current/pgbench.html). Under the hood all benchmarks are run by pgbench.

DBench parses the result of each benchmarks and persists it in a database. This allows for easy comparisons of different benchmarks. The endgoal is for DBench to automatically generate multiple insightful plots that help with hunting down potential performance culprits.

## Pre-requisites <a id="prerequisites"></a>

- [PostgreSQL](https://www.postgresql.org/) (we need a database to benchmark against)
- [pgbench](https://www.postgresql.org/docs/current/pgbench.html) (the actual benchmarking tool)
- [gnuplot](http://www.gnuplot.info/) (for plotting the results)

## Install <a id="install"></a>

> Important: While the releases offer binaries for multiple platforms and architectures, only Linux is tested. If you are using a different OS, I do not guarantee that dbench will work as expected.

Download one of the [releases](https://github.com/nikoksr/dbench/releases) for your system, or install using the provided [install script](scripts/install.sh):

```sh
curl -L https://tinyurl.com/install-dbench | bash
```

Alternatively, you can install dbench using Go:

```sh
go install github.com/nikoksr/dbench
```

## Usage <a id="usage"></a>

> It is recommended to check the help page of the command line interface for more information on the available flags and commands.

Before you can run any benchmarks, you need to create a database and initialize it with pgbench. This can be done by running the following command:

> Hint: Remember to replace the flags with your own values.

```sh
dbench bench init --dbname postgres --username postgres --host 127.0.0.1 --port 5432
```

> Hint: dbench/pgbench expects the `PGPASSWORD` environment variable to be set. Currently no password flag is supported since I didn't need it and it enforces better security practices. This might very well change down the line.

Now, you can run your first benchmark using the following command:

```sh
dbench bench run --dbname postgres --username postgres --host 127.0.0.1 --port 5432
```

The benchmark will present you with an executable command once it is done. You can use this command to generate a plot of the results. It looks something like this:

```sh
dbench plot <benckmark-id>
```

Under the hood we generate gnuplot compatible data fields and a gnuplot script. The script is then executed and the plot is generated. The plot is saved in the current working directory as a PNG.

To check on old benchmarks, you can use the `list` command:

```sh
dbench bench list
```
