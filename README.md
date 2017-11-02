# From CSV to Buxfer
This is a personal experiment where I'm pushing my personal economics data (CSV) to [Buxfer](https://www.buxfer.com) (through Buxfer APIs)

## Requirements

 - docker (>= 17.06.0-ce)
 - docker-compose (>= 1.14.0)

## Install
Few simple steps:

1. clone
```bash
$ git clone https://github.com/wilk/economics-data-collector.git
```

2. setup python
```bash
$ docker-compose run --rm setup-python
```

2. setup golang
```bash
$ docker-compose run --rm setup-golang
```

## Usage
1. cleaning data
```bash
$ docker-compose run --rm cleaner
```

2. collecting data
```bash
$ docker-compose run --rm collector
```

3. pushing data
```bash
$ docker-compose run --rm goxfer
```