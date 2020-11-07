# SURGARD client

### Protocol description

`5PPR MTAAAAQXYZGGCCC` where:

- `5` - surgrard flag, it is always `5`
- `PP` - reciever number 01-99
- `R` - reciever line number 1-9
- `MT` - message type, 12(preferably) or 98
- `AAAA` - 4 symbols abonent number
- `Q` - type of event, `1` or `E` - new event, `3` or `R` - close/cancel
- `XYZ` - event code 001-999
- `GG` - group or section number 00-99
- `CCC` - zone/key number 000-999

### Install

```sh
go get github.com/sg3des/surgard
```

### Usage

```go
c, err := surgrard.NewClient(addr)
if err != nil {...}

c.Dial(surgard.DialData{
	PP: 01,
	R: 01,
	Object: 1234,
	Close: false,
	Code: 130,
	Group: 01,
	CCC: 001,
})
```
