# P0 Bug 修复记录

## Bug 1: EXPECTED_CRLF
- 文件：`C:\Users\i\.openclaw\workspace\gostalkd\net\prot.go`
- 修改内容：修复 `dispatchOpPut` 中 CRLF 检测逻辑
- 修复前：`if c.Cmd[idx] != '\r'`
- 修复后：`if c.Cmd[idx] != '\r' || c.Cmd[idx+1] != '\n'`

## Bug 2: reserve-job NOT_FOUND
- 文件：`C:\Users\i\.openclaw\workspace\gostalkd\net\prot.go`
- 修改内容：修复 `dispatchOpReserveJob` 中对 Reserved 状态 job 的处理
- 修复前：switch 只处理 Ready/Buried/Delayed，Reserved 状态落入 default 返回 INTERNAL_ERROR
- 修复后：增加 `case constant.Reserved: replyMsg(c, constant.MsgNotFound); return`
