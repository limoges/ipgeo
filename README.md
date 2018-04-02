# ipgeo: sample IP based geolocation application

From Wikipedia, 'Geolocation':
> Geolocation is the identification or estimation of the real-world geographic 
> location of an object, such as a radar source, mobile phone, or computer.

The process is not exact, since ip addresses are not distributed based on
geographic location. It can be used as a decent approximation though, because
of the somewhat stable nature of telecommunication networks used in conjunction
with isp, domains, user-sourced gps coordinates, etc.

# Aim

The goal of this project is to provide a coding sample to examine.

# Goal

Implementing an http service which implements ip-based geolocation that works
like the following snippet:

```
$ curl http://localhost:8080/ip/87.168.167.128

200/ok

{
        "country": "Germany",
        "subdivision1": "Land Berlin",
        "subdivision2": "",
        "city": "Berlin",
}
```

To achieve this, two CSV files found in `data/` are provided; ipv4.csv contains
the IPv4 networks in CIDR format and their location IDs. locations.csv contains
the location IDs and their related data (country, subdivision1, ...).

