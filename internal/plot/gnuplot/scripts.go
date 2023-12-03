package gnuplot

var ScriptTemplates = map[string]string{
	"overview":                                    overview,
	"performance_and_resource_correlation":        performanceAndResourceCorrelation,
	"latency_and_system_load_over_clients":        latencyAndSystemLoadOverClients,
	"cpu_load_distribution":                       cpuLoadDistribution,
	"memory_load_distribution":                    memoryLoadDistribution,
	"performance_efficiency":                      performanceEfficiency,
	"transactions_latency_conn_time_over_clients": transactionsLatencyConnTimeOverClients,
}

const (
	overview = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,800 enhanced font 'Verdana,10'
set datafile separator ","
set multiplot layout 4, 1 title "Comprehensive Benchmark Overview"

set xlabel "Number of Clients"
set autoscale y

set ylabel "Transactions Per Second"
set key bottom right
plot '{{ .DataPath }}' using "Clients":"TransactionsPerSecond" title "Transactions Per Second" with lines

set ylabel "Average Latency"
set key bottom right
plot '{{ .DataPath }}' using "Clients":"AverageLatency" title "Average Latency" with lines

set ylabel "CPU Average Load"
set key bottom right
plot '{{ .DataPath }}' using "Clients":"CPUAverageLoad" title "CPU Average Load" with lines

set ylabel "Memory Average Load"
set key bottom right
plot '{{ .DataPath }}' using "Clients":"MemoryAverageLoad" title "Memory Average Load" with lines

unset multiplot
`

	performanceAndResourceCorrelation = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "Performance and Resource Load Correlation"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "Transactions Per Second"
set y2label "CPU and Memory Load"
set y2tics
set grid
plot '{{ .DataPath }}' using "Clients":"TransactionsPerSecond" title "Transactions Per Second" with lines, \
     '{{ .DataPath }}' using "Clients":"CPUAverageLoad" title "CPU Average Load" axes x1y2 with lines, \
     '{{ .DataPath }}' using "Clients":"MemoryAverageLoad" title "Memory Average Load" axes x1y2 with lines
`

	latencyAndSystemLoadOverClients = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "Latency and System Load Over Clients"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "Average Latency (ms)"
set y2label "CPU 95th Load and Memory 95th Load"
set xtics auto
set ytics auto
set y2tics auto
set grid
plot '{{ .DataPath }}' using "Clients":"AverageLatency" title "Average Latency" with lines, \
     '{{ .DataPath }}' using "Clients":"CPU95thLoad" title "CPU 95th Load" axes x1y2 with lines, \
     '{{ .DataPath }}' using "Clients":"Memory95thLoad" title "Memory 95th Load" axes x1y2 with lines
`

	cpuLoadDistribution = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "CPU Load Distribution"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "CPU Load (%)"
set grid
plot '{{ .DataPath }}' using "Clients":"CPUMinLoad" title "Min Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"CPUMaxLoad" title "Max Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"CPU50thLoad" title "50th Load" with lines, \
     '{{ .DataPath }}' using "Clients":"CPU75thLoad" title "75th Load" with lines, \
     '{{ .DataPath }}' using "Clients":"CPU95thLoad" title "95th Load" with lines, \
     '{{ .DataPath }}' using "Clients":"CPU99thLoad" title "99th Load" with lines
`

	memoryLoadDistribution = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "Memory Load Distribution"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "Memory Load (%)"
set grid
plot '{{ .DataPath }}' using "Clients":"MemoryMinLoad" title "Min Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"MemoryMaxLoad" title "Max Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"Memory50thLoad" title "50th Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"Memory75thLoad" title "75th Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"Memory95thLoad" title "95th Load" with lines, \
	 '{{ .DataPath }}' using "Clients":"Memory99thLoad" title "99th Load" with lines
`

	performanceEfficiency = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "Performance Efficiency"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "Transactions Per Second per CPU Average Load"
set grid
plot '{{ .DataPath }}' using "Clients":(column("TransactionsPerSecond")/column("CPUAverageLoad")) title "Transactions/sec per CPU Load" with lines
`

	transactionsLatencyConnTimeOverClients = `set datafile separator ","
set output '{{ .OutputPath }}'
set terminal pngcairo size 800,600 enhanced font 'Verdana,10'
set title "Transactions per Second, Connection Time and Average Latency over Clients"
set key bottom right
set tmargin 5
set xlabel "Number of Clients"
set ylabel "Transactions per Second"
set y2label "Connection Time and Average Latency"
set ytics nomirror
set y2tics
set format y2 "%.0fms"
set autoscale y
set autoscale y2
set grid

plot '{{ .DataPath }}' using "Clients":"TransactionsPerSecond" with linespoints title "TPS", \
	 '{{ .DataPath }}' using "Clients":"ConnectionTime" with linespoints title "Connection Time" axes x1y2, \
	 '{{ .DataPath }}' using "Clients":"AverageLatency" with linespoints title "Latency" axes x1y2
`
)
