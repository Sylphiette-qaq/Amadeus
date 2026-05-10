# 数据模型

<h2 id="tocS_BaseResponse">BaseResponse</h2>

<a id="schemabaseresponse"></a>
<a id="schema_BaseResponse"></a>
<a id="tocSbaseresponse"></a>
<a id="tocsbaseresponse"></a>

```json
{
  "status": "string",
  "retcode": 0,
  "data": null,
  "message": "string",
  "wording": "string",
  "stream": "stream-action"
}

```

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|status|string|true|none||状态 (ok/failed)|
|retcode|number|true|none||返回码|
|data|any|false|none||业务数据（具体结构由各接口定义）|
|message|string|false|none||消息|
|wording|string|false|none||提示|
|stream|string|false|none||流式响应|

#### 枚举值

|属性|值|
|---|---|
|stream|stream-action|
|stream|normal-action|

<h2 id="tocS_EmptyData">EmptyData</h2>

<a id="schemaemptydata"></a>
<a id="schema_EmptyData"></a>
<a id="tocSemptydata"></a>
<a id="tocsemptydata"></a>

```json
null

```

无数据

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|null|false|none||无数据|

<h2 id="tocS_FileBaseData">FileBaseData</h2>

<a id="schemafilebasedata"></a>
<a id="schema_FileBaseData"></a>
<a id="tocSfilebasedata"></a>
<a id="tocsfilebasedata"></a>

```json
{
  "file": "string",
  "path": "string",
  "url": "string",
  "name": "string",
  "thumb": "string"
}

```

文件消息段基础数据

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|file|string|true|none||文件路径/URL/file:///|
|path|string|false|none||文件路径|
|url|string|false|none||文件URL|
|name|string|false|none||文件名|
|thumb|string|false|none||缩略图|

<h2 id="tocS_OB11MessageAt">OB11MessageAt</h2>

<a id="schemaob11messageat"></a>
<a id="schema_OB11MessageAt"></a>
<a id="tocSob11messageat"></a>
<a id="tocsob11messageat"></a>

```json
{
  "type": "at",
  "data": {
    "qq": "string",
    "name": "string"
  }
}

```

@消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» qq|string|true|none||QQ号或all|
|» name|string|false|none||显示名称|

#### 枚举值

|属性|值|
|---|---|
|type|at|

<h2 id="tocS_OB11MessageContact">OB11MessageContact</h2>

<a id="schemaob11messagecontact"></a>
<a id="schema_OB11MessageContact"></a>
<a id="tocSob11messagecontact"></a>
<a id="tocsob11messagecontact"></a>

```json
{
  "type": "contact",
  "data": {
    "type": "qq",
    "id": "string"
  }
}

```

联系人消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» type|string|true|none||联系人类型|
|» id|string|true|none||联系人ID|

#### 枚举值

|属性|值|
|---|---|
|type|contact|
|type|qq|
|type|group|

<h2 id="tocS_OB11MessageCustomMusic">OB11MessageCustomMusic</h2>

<a id="schemaob11messagecustommusic"></a>
<a id="schema_OB11MessageCustomMusic"></a>
<a id="tocSob11messagecustommusic"></a>
<a id="tocsob11messagecustommusic"></a>

```json
{
  "type": "music",
  "data": {
    "type": "qq",
    "id": null,
    "url": "string",
    "audio": "string",
    "title": "string",
    "image": "string",
    "content": "string"
  }
}

```

自定义音乐消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» type|string|true|none||音乐平台类型|
|» id|null|true|none||none|
|» url|string|true|none||点击后跳转URL|
|» audio|string|false|none||音频URL|
|» title|string|false|none||音乐标题|
|» image|string|true|none||封面图片URL|
|» content|string|false|none||音乐简介|

#### 枚举值

|属性|值|
|---|---|
|type|music|
|type|qq|
|type|163|
|type|kugou|
|type|migu|
|type|kuwo|
|type|custom|

<h2 id="tocS_OB11MessageData">OB11MessageData</h2>

<a id="schemaob11messagedata"></a>
<a id="schema_OB11MessageData"></a>
<a id="tocSob11messagedata"></a>
<a id="tocsob11messagedata"></a>

```json
{
  "type": "text",
  "data": {
    "text": "string"
  }
}

```

OneBot 11 消息段

### 属性

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageText](#schemaob11messagetext)|false|none||纯文本消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageFace](#schemaob11messageface)|false|none||QQ表情消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageMFace](#schemaob11messagemface)|false|none||商城表情消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageAt](#schemaob11messageat)|false|none||@消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageReply](#schemaob11messagereply)|false|none||回复消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageImage](#schemaob11messageimage)|false|none||图片消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageRecord](#schemaob11messagerecord)|false|none||语音消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageVideo](#schemaob11messagevideo)|false|none||视频消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageFile](#schemaob11messagefile)|false|none||文件消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageIdMusic](#schemaob11messageidmusic)|false|none||ID音乐消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageCustomMusic](#schemaob11messagecustommusic)|false|none||自定义音乐消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessagePoke](#schemaob11messagepoke)|false|none||戳一戳消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageDice](#schemaob11messagedice)|false|none||骰子消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageRPS](#schemaob11messagerps)|false|none||猜拳消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageContact](#schemaob11messagecontact)|false|none||联系人消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageLocation](#schemaob11messagelocation)|false|none||位置消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageJson](#schemaob11messagejson)|false|none||JSON消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageXml](#schemaob11messagexml)|false|none||XML消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageMarkdown](#schemaob11messagemarkdown)|false|none||Markdown消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageMiniApp](#schemaob11messageminiapp)|false|none||小程序消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageNode](#schemaob11messagenode)|false|none||合并转发消息节点|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageForward](#schemaob11messageforward)|false|none||合并转发消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageOnlineFile](#schemaob11messageonlinefile)|false|none||在线文件消息段|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageFlashTransfer](#schemaob11messageflashtransfer)|false|none||QQ闪传消息段|

<h2 id="tocS_OB11MessageDice">OB11MessageDice</h2>

<a id="schemaob11messagedice"></a>
<a id="schema_OB11MessageDice"></a>
<a id="tocSob11messagedice"></a>
<a id="tocsob11messagedice"></a>

```json
{
  "type": "dice",
  "data": {
    "result": 0
  }
}

```

骰子消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» result|any|true|none||骰子结果|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|type|dice|

<h2 id="tocS_OB11MessageFace">OB11MessageFace</h2>

<a id="schemaob11messageface"></a>
<a id="schema_OB11MessageFace"></a>
<a id="tocSob11messageface"></a>
<a id="tocsob11messageface"></a>

```json
{
  "type": "face",
  "data": {
    "id": "string",
    "resultId": "string",
    "chainCount": 0
  }
}

```

QQ表情消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» id|string|true|none||表情ID|
|» resultId|string|false|none||结果ID|
|» chainCount|number|false|none||连击数|

#### 枚举值

|属性|值|
|---|---|
|type|face|

<h2 id="tocS_OB11MessageFileBase">OB11MessageFileBase</h2>

<a id="schemaob11messagefilebase"></a>
<a id="schema_OB11MessageFileBase"></a>
<a id="tocSob11messagefilebase"></a>
<a id="tocsob11messagefilebase"></a>

```json
{
  "data": {
    "file": "string",
    "path": "string",
    "url": "string",
    "name": "string",
    "thumb": "string"
  }
}

```

文件消息基础接口

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|data|[FileBaseData](#schemafilebasedata)|true|none||文件消息段基础数据|

<h2 id="tocS_OB11MessageFile">OB11MessageFile</h2>

<a id="schemaob11messagefile"></a>
<a id="schema_OB11MessageFile"></a>
<a id="tocSob11messagefile"></a>
<a id="tocsob11messagefile"></a>

```json
{
  "type": "file",
  "data": {
    "file": "string",
    "path": "string",
    "url": "string",
    "name": "string",
    "thumb": "string"
  }
}

```

文件消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|[FileBaseData](#schemafilebasedata)|true|none||文件消息段基础数据|

#### 枚举值

|属性|值|
|---|---|
|type|file|

<h2 id="tocS_OB11MessageFlashTransfer">OB11MessageFlashTransfer</h2>

<a id="schemaob11messageflashtransfer"></a>
<a id="schema_OB11MessageFlashTransfer"></a>
<a id="tocSob11messageflashtransfer"></a>
<a id="tocsob11messageflashtransfer"></a>

```json
{
  "type": "flashtransfer",
  "data": {
    "fileSetId": "string"
  }
}

```

QQ闪传消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» fileSetId|string|true|none||文件集ID|

#### 枚举值

|属性|值|
|---|---|
|type|flashtransfer|

<h2 id="tocS_OB11MessageForward">OB11MessageForward</h2>

<a id="schemaob11messageforward"></a>
<a id="schema_OB11MessageForward"></a>
<a id="tocSob11messageforward"></a>
<a id="tocsob11messageforward"></a>

```json
{
  "type": "forward",
  "data": {
    "id": "string",
    "content": {}
  }
}

```

合并转发消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» id|string|true|none||合并转发ID|
|» content|object|false|none||消息内容 (OB11Message[])|

#### 枚举值

|属性|值|
|---|---|
|type|forward|

<h2 id="tocS_OB11MessageIdMusic">OB11MessageIdMusic</h2>

<a id="schemaob11messageidmusic"></a>
<a id="schema_OB11MessageIdMusic"></a>
<a id="tocSob11messageidmusic"></a>
<a id="tocsob11messageidmusic"></a>

```json
{
  "type": "music",
  "data": {
    "type": "qq",
    "id": "string"
  }
}

```

ID音乐消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» type|string|true|none||音乐平台类型|
|» id|any|true|none||音乐ID|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|type|music|
|type|qq|
|type|163|
|type|kugou|
|type|migu|
|type|kuwo|

<h2 id="tocS_OB11MessageImage">OB11MessageImage</h2>

<a id="schemaob11messageimage"></a>
<a id="schema_OB11MessageImage"></a>
<a id="tocSob11messageimage"></a>
<a id="tocsob11messageimage"></a>

```json
{
  "type": "image",
  "data": {
    "file": "string",
    "path": "string",
    "url": "string",
    "name": "string",
    "thumb": "string",
    "summary": "string",
    "sub_type": 0
  }
}

```

图片消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|any|true|none||none|

allOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|object|false|none||文件消息段基础数据|
|»» file|string|true|none||文件路径/URL/file:///|
|»» path|string|false|none||文件路径|
|»» url|string|false|none||文件URL|
|»» name|string|false|none||文件名|
|»» thumb|string|false|none||缩略图|

and

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|object|false|none||none|
|»» summary|string|false|none||图片摘要|
|»» sub_type|number|false|none||图片子类型|

#### 枚举值

|属性|值|
|---|---|
|type|image|

<h2 id="tocS_OB11MessageJson">OB11MessageJson</h2>

<a id="schemaob11messagejson"></a>
<a id="schema_OB11MessageJson"></a>
<a id="tocSob11messagejson"></a>
<a id="tocsob11messagejson"></a>

```json
{
  "type": "json",
  "data": {
    "data": "string",
    "config": {
      "token": "string"
    }
  }
}

```

JSON消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» data|any|true|none||JSON数据|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|object|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» config|object|false|none||none|
|»» token|string|true|none||token|

#### 枚举值

|属性|值|
|---|---|
|type|json|

<h2 id="tocS_OB11MessageLocation">OB11MessageLocation</h2>

<a id="schemaob11messagelocation"></a>
<a id="schema_OB11MessageLocation"></a>
<a id="tocSob11messagelocation"></a>
<a id="tocsob11messagelocation"></a>

```json
{
  "type": "location",
  "data": {
    "lat": "string",
    "lon": "string",
    "title": "string",
    "content": "string"
  }
}

```

位置消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» lat|any|true|none||纬度|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» lon|any|true|none||经度|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» title|string|false|none||标题|
|» content|string|false|none||内容|

#### 枚举值

|属性|值|
|---|---|
|type|location|

<h2 id="tocS_OB11MessageMFace">OB11MessageMFace</h2>

<a id="schemaob11messagemface"></a>
<a id="schema_OB11MessageMFace"></a>
<a id="tocSob11messagemface"></a>
<a id="tocsob11messagemface"></a>

```json
{
  "type": "mface",
  "data": {
    "emoji_package_id": 0,
    "emoji_id": "string",
    "key": "string",
    "summary": "string"
  }
}

```

商城表情消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» emoji_package_id|number|true|none||表情包ID|
|» emoji_id|string|true|none||表情ID|
|» key|string|true|none||表情key|
|» summary|string|true|none||表情摘要|

#### 枚举值

|属性|值|
|---|---|
|type|mface|

<h2 id="tocS_OB11MessageMarkdown">OB11MessageMarkdown</h2>

<a id="schemaob11messagemarkdown"></a>
<a id="schema_OB11MessageMarkdown"></a>
<a id="tocSob11messagemarkdown"></a>
<a id="tocsob11messagemarkdown"></a>

```json
{
  "type": "markdown",
  "data": {
    "content": "string"
  }
}

```

Markdown消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» content|string|true|none||Markdown内容|

#### 枚举值

|属性|值|
|---|---|
|type|markdown|

<h2 id="tocS_OB11MessageMiniApp">OB11MessageMiniApp</h2>

<a id="schemaob11messageminiapp"></a>
<a id="schema_OB11MessageMiniApp"></a>
<a id="tocSob11messageminiapp"></a>
<a id="tocsob11messageminiapp"></a>

```json
{
  "type": "miniapp",
  "data": {
    "data": "string"
  }
}

```

小程序消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» data|string|true|none||小程序数据|

#### 枚举值

|属性|值|
|---|---|
|type|miniapp|

<h2 id="tocS_OB11MessageMixType">OB11MessageMixType</h2>

<a id="schemaob11messagemixtype"></a>
<a id="schema_OB11MessageMixType"></a>
<a id="tocSob11messagemixtype"></a>
<a id="tocsob11messagemixtype"></a>

```json
[
  {
    "type": "text",
    "data": {
      "text": "string"
    }
  }
]

```

OneBot 11 消息混合类型

### 属性

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[[OB11MessageData](#schemaob11messagedata)]|false|none||[OneBot 11 消息段]|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|string|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|[OB11MessageData](#schemaob11messagedata)|false|none||OneBot 11 消息段|

<h2 id="tocS_OB11MessageNode">OB11MessageNode</h2>

<a id="schemaob11messagenode"></a>
<a id="schema_OB11MessageNode"></a>
<a id="tocSob11messagenode"></a>
<a id="tocsob11messagenode"></a>

```json
{
  "type": "node",
  "data": {
    "id": "string",
    "user_id": 0,
    "uin": 0,
    "nickname": "string",
    "name": "string",
    "content": {},
    "source": "string",
    "news": [
      {
        "text": "string"
      }
    ],
    "summary": "string",
    "prompt": "string",
    "time": "string"
  }
}

```

合并转发消息节点

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» id|string|false|none||转发消息ID|
|» user_id|any|false|none||发送者QQ号|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» uin|any|false|none||发送者QQ号(兼容go-cqhttp)|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» nickname|string|true|none||发送者昵称|
|» name|string|false|none||发送者昵称(兼容go-cqhttp)|
|» content|object|true|none||消息内容 (OB11MessageMixType)|
|» source|string|false|none||消息来源|
|» news|[object]|false|none||none|
|»» text|string|true|none||新闻文本|
|» summary|string|false|none||摘要|
|» prompt|string|false|none||提示|
|» time|string|false|none||时间|

#### 枚举值

|属性|值|
|---|---|
|type|node|

<h2 id="tocS_OB11MessageOnlineFile">OB11MessageOnlineFile</h2>

<a id="schemaob11messageonlinefile"></a>
<a id="schema_OB11MessageOnlineFile"></a>
<a id="tocSob11messageonlinefile"></a>
<a id="tocsob11messageonlinefile"></a>

```json
{
  "type": "onlinefile",
  "data": {
    "msgId": "string",
    "elementId": "string",
    "fileName": "string",
    "fileSize": "string",
    "isDir": true
  }
}

```

在线文件消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» msgId|string|true|none||消息ID|
|» elementId|string|true|none||元素ID|
|» fileName|string|true|none||文件名|
|» fileSize|string|true|none||文件大小|
|» isDir|boolean|true|none||是否为目录|

#### 枚举值

|属性|值|
|---|---|
|type|onlinefile|

<h2 id="tocS_OB11MessagePoke">OB11MessagePoke</h2>

<a id="schemaob11messagepoke"></a>
<a id="schema_OB11MessagePoke"></a>
<a id="tocSob11messagepoke"></a>
<a id="tocsob11messagepoke"></a>

```json
{
  "type": "poke",
  "data": {
    "type": "string",
    "id": "string"
  }
}

```

戳一戳消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» type|string|true|none||戳一戳类型|
|» id|string|true|none||戳一戳ID|

#### 枚举值

|属性|值|
|---|---|
|type|poke|

<h2 id="tocS_OB11MessageRPS">OB11MessageRPS</h2>

<a id="schemaob11messagerps"></a>
<a id="schema_OB11MessageRPS"></a>
<a id="tocSob11messagerps"></a>
<a id="tocsob11messagerps"></a>

```json
{
  "type": "rps",
  "data": {
    "result": 0
  }
}

```

猜拳消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» result|any|true|none||猜拳结果|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» *anonymous*|string|false|none||none|

#### 枚举值

|属性|值|
|---|---|
|type|rps|

<h2 id="tocS_OB11MessageRecord">OB11MessageRecord</h2>

<a id="schemaob11messagerecord"></a>
<a id="schema_OB11MessageRecord"></a>
<a id="tocSob11messagerecord"></a>
<a id="tocsob11messagerecord"></a>

```json
{
  "type": "record",
  "data": {
    "file": "string",
    "path": "string",
    "url": "string",
    "name": "string",
    "thumb": "string"
  }
}

```

语音消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|[FileBaseData](#schemafilebasedata)|true|none||文件消息段基础数据|

#### 枚举值

|属性|值|
|---|---|
|type|record|

<h2 id="tocS_OB11MessageReply">OB11MessageReply</h2>

<a id="schemaob11messagereply"></a>
<a id="schema_OB11MessageReply"></a>
<a id="tocSob11messagereply"></a>
<a id="tocsob11messagereply"></a>

```json
{
  "type": "reply",
  "data": {
    "id": "string",
    "seq": 0
  }
}

```

回复消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» id|string|false|none||消息ID的短ID映射|
|» seq|number|false|none||消息序列号，优先使用|

#### 枚举值

|属性|值|
|---|---|
|type|reply|

<h2 id="tocS_OB11Message">OB11Message</h2>

<a id="schemaob11message"></a>
<a id="schema_OB11Message"></a>
<a id="tocSob11message"></a>
<a id="tocsob11message"></a>

```json
{
  "real_seq": "string",
  "temp_source": 0,
  "message_sent_type": "string",
  "target_id": 0,
  "self_id": 0,
  "time": 0,
  "message_id": 0,
  "message_seq": 0,
  "real_id": 0,
  "user_id": 0,
  "group_id": 0,
  "group_name": "string",
  "message_type": "private",
  "sub_type": "friend",
  "sender": {
    "user_id": 0,
    "nickname": "string",
    "card": "string",
    "role": "string",
    "sex": "string",
    "age": 0,
    "area": "string",
    "level": "string",
    "title": "string"
  },
  "message": [
    {
      "type": "[",
      "data": {}
    }
  ],
  "message_format": "array",
  "raw_message": "string",
  "font": 0,
  "post_type": "string",
  "raw": {},
  "emoji_likes_list": [
    {
      "emoji_id": "string",
      "emoji_type": "string",
      "likes_cnt": "string"
    }
  ]
}

```

OneBot 11 完整消息对象

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|real_seq|string|false|none||真实序列号|
|temp_source|number|false|none||临时会话来源|
|message_sent_type|string|false|none||消息发送类型|
|target_id|number|false|none||目标ID|
|self_id|number|false|none||机器人QQ号|
|time|number|true|none||消息时间戳|
|message_id|number|true|none||消息ID|
|message_seq|number|true|none||消息序列号|
|real_id|number|true|none||真实ID|
|user_id|any|true|none||发送者QQ号|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|group_id|any|false|none||群号|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|group_name|string|false|none||群名称|
|message_type|string|true|none||消息类型|
|sub_type|string|false|none||消息子类型|
|sender|[OB11Sender](#schemaob11sender)|true|none||OneBot 11 发送者信息|
|message|any|true|none||消息内容|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|[[OB11MessageData](#schemaob11messagedata)]|false|none||[OneBot 11 消息段]|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|message_format|string|true|none||消息格式|
|raw_message|string|true|none||原始消息|
|font|number|true|none||字体|
|post_type|string|false|none||上报类型|
|raw|object|false|none||原始消息对象|
|emoji_likes_list|[object]|false|none||表情点赞列表|
|» emoji_id|string|true|none||表情ID|
|» emoji_type|string|true|none||表情类型|
|» likes_cnt|string|true|none||点赞数|

#### 枚举值

|属性|值|
|---|---|
|message_type|private|
|message_type|group|
|sub_type|friend|
|sub_type|group|
|sub_type|normal|
|message_format|array|
|message_format|string|

<h2 id="tocS_OB11MessageText">OB11MessageText</h2>

<a id="schemaob11messagetext"></a>
<a id="schema_OB11MessageText"></a>
<a id="tocSob11messagetext"></a>
<a id="tocsob11messagetext"></a>

```json
{
  "type": "text",
  "data": {
    "text": "string"
  }
}

```

纯文本消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» text|string|true|none||纯文本内容|

#### 枚举值

|属性|值|
|---|---|
|type|text|

<h2 id="tocS_OB11MessageVideo">OB11MessageVideo</h2>

<a id="schemaob11messagevideo"></a>
<a id="schema_OB11MessageVideo"></a>
<a id="tocSob11messagevideo"></a>
<a id="tocsob11messagevideo"></a>

```json
{
  "type": "video",
  "data": {
    "file": "string",
    "path": "string",
    "url": "string",
    "name": "string",
    "thumb": "string"
  }
}

```

视频消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|[FileBaseData](#schemafilebasedata)|true|none||文件消息段基础数据|

#### 枚举值

|属性|值|
|---|---|
|type|video|

<h2 id="tocS_OB11MessageXml">OB11MessageXml</h2>

<a id="schemaob11messagexml"></a>
<a id="schema_OB11MessageXml"></a>
<a id="tocSob11messagexml"></a>
<a id="tocsob11messagexml"></a>

```json
{
  "type": "xml",
  "data": {
    "data": "string"
  }
}

```

XML消息段

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|type|string|true|none||none|
|data|object|true|none||none|
|» data|string|true|none||XML数据|

#### 枚举值

|属性|值|
|---|---|
|type|xml|

<h2 id="tocS_OB11PostSendMsg">OB11PostSendMsg</h2>

<a id="schemaob11postsendmsg"></a>
<a id="schema_OB11PostSendMsg"></a>
<a id="tocSob11postsendmsg"></a>
<a id="tocsob11postsendmsg"></a>

```json
{
  "message_type": "private",
  "user_id": "string",
  "group_id": "string",
  "message": [
    {
      "type": "[",
      "data": {}
    }
  ],
  "messages": [
    {
      "type": "[",
      "data": {}
    }
  ],
  "auto_escape": true,
  "source": "string",
  "news": [
    {
      "text": "string"
    }
  ],
  "summary": "string",
  "prompt": "string",
  "time": "string"
}

```

OneBot 11 发送消息请求

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|message_type|string|false|none||消息类型|
|user_id|string|false|none||用户QQ号|
|group_id|string|false|none||群号|
|message|[OB11MessageMixType](#schemaob11messagemixtype)|true|none||OneBot 11 消息混合类型|
|messages|[OB11MessageMixType](#schemaob11messagemixtype)|false|none||OneBot 11 消息混合类型|
|auto_escape|any|false|none||是否作为纯文本发送|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|boolean|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|source|string|false|none||消息来源|
|news|[object]|false|none||none|
|» text|string|true|none||文本|
|summary|string|false|none||摘要|
|prompt|string|false|none||提示|
|time|string|false|none||时间|

#### 枚举值

|属性|值|
|---|---|
|message_type|private|
|message_type|group|

<h2 id="tocS_OB11Sender">OB11Sender</h2>

<a id="schemaob11sender"></a>
<a id="schema_OB11Sender"></a>
<a id="tocSob11sender"></a>
<a id="tocsob11sender"></a>

```json
{
  "user_id": 0,
  "nickname": "string",
  "card": "string",
  "role": "string",
  "sex": "string",
  "age": 0,
  "area": "string",
  "level": "string",
  "title": "string"
}

```

OneBot 11 发送者信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|user_id|any|true|none||发送者QQ号|

anyOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|number|false|none||none|

or

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||none|

continued

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|nickname|string|true|none||发送者昵称|
|card|string|false|none||群名片|
|role|string|false|none||角色|
|sex|string|false|none||性别|
|age|number|false|none||年龄|
|area|string|false|none||地区|
|level|string|false|none||等级|
|title|string|false|none||头衔|

<h2 id="tocS_OB11GroupMember">OB11GroupMember</h2>

<a id="schemaob11groupmember"></a>
<a id="schema_OB11GroupMember"></a>
<a id="tocSob11groupmember"></a>
<a id="tocsob11groupmember"></a>

```json
{
  "group_id": 0,
  "user_id": 0,
  "nickname": "string",
  "card": "string",
  "sex": "string",
  "age": 0,
  "join_time": 0,
  "last_sent_time": 0,
  "level": "string",
  "qq_level": 0,
  "role": "string",
  "title": "string",
  "area": "string",
  "unfriendly": true,
  "title_expire_time": 0,
  "card_changeable": true,
  "shut_up_timestamp": 0,
  "is_robot": true,
  "qage": 0
}

```

OneBot 11 群成员信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|group_id|number|true|none||群号|
|user_id|number|true|none||QQ号|
|nickname|string|true|none||昵称|
|card|string|false|none||名片|
|sex|string|false|none||性别|
|age|number|false|none||年龄|
|join_time|number|false|none||入群时间戳|
|last_sent_time|number|false|none||最后发言时间戳|
|level|string|false|none||等级|
|qq_level|number|false|none||QQ等级|
|role|string|false|none||角色 (owner/admin/member)|
|title|string|false|none||头衔|
|area|string|false|none||地区|
|unfriendly|boolean|false|none||是否不良记录|
|title_expire_time|number|false|none||头衔过期时间|
|card_changeable|boolean|false|none||是否允许修改名片|
|shut_up_timestamp|number|false|none||禁言截止时间戳|
|is_robot|boolean|false|none||是否为机器人|
|qage|number|false|none||Q龄|

<h2 id="tocS_OB11Group">OB11Group</h2>

<a id="schemaob11group"></a>
<a id="schema_OB11Group"></a>
<a id="tocSob11group"></a>
<a id="tocsob11group"></a>

```json
{
  "group_all_shut": 0,
  "group_remark": "string",
  "group_id": 0,
  "group_name": "string",
  "member_count": 0,
  "max_member_count": 0
}

```

OneBot 11 群信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|group_all_shut|number|true|none||是否全员禁言|
|group_remark|string|true|none||群备注|
|group_id|number|true|none||群号|
|group_name|string|true|none||群名称|
|member_count|number|false|none||成员人数|
|max_member_count|number|false|none||最大成员人数|

<h2 id="tocS_OB11ActionMessage">OB11ActionMessage</h2>

<a id="schemaob11actionmessage"></a>
<a id="schema_OB11ActionMessage"></a>
<a id="tocSob11actionmessage"></a>
<a id="tocsob11actionmessage"></a>

```json
{
  "self_id": 0,
  "user_id": 0,
  "time": 0,
  "real_seq": "string",
  "message_type": "string",
  "sender": {
    "user_id": 0,
    "nickname": "string",
    "card": "string",
    "role": "string"
  },
  "raw_message": "string",
  "font": 0,
  "sub_type": "string",
  "message": {},
  "message_format": "string",
  "post_type": "string",
  "group_id": 0,
  "group_name": "string",
  "message_id": 0,
  "message_seq": 0,
  "emoji_likes_list": [
    {
      "emoji_id": "string",
      "emoji_type": "string",
      "likes_cnt": "string"
    }
  ]
}

```

OneBot 11 消息信息

### 属性

allOf

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|object|false|none||最后一条消息|
|» self_id|number|true|none||发送者QQ号|
|» user_id|number|true|none||接收者QQ号|
|» time|number|true|none||时间戳|
|» real_seq|string|true|none||消息序号|
|» message_type|string|true|none||消息类型|
|» sender|object|true|none||none|
|»» user_id|number|true|none||用户QQ号|
|»» nickname|string|true|none||用户昵称|
|»» card|string|false|none||用户名片|
|»» role|string|false|none||用户角色|
|» raw_message|string|true|none||原始消息|
|» font|number|true|none||字体大小|
|» sub_type|string|true|none||子类型|
|» message|object|true|none||消息内容|
|» message_format|string|true|none||消息格式|
|» post_type|string|true|none||发布类型|
|» group_id|number|true|none||群号|
|» group_name|string|true|none||群名称|

and

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|*anonymous*|object|false|none||OneBot 11 消息信息|
|» message_id|number|true|none||消息ID|
|» message_seq|number|true|none||消息序列号|
|» emoji_likes_list|[object]|true|none||none|
|»» emoji_id|string|true|none||表情符号ID|
|»» emoji_type|string|true|none||表情符号类型|
|»» likes_cnt|string|true|none||点赞数|

<h2 id="tocS_OB11Notify">OB11Notify</h2>

<a id="schemaob11notify"></a>
<a id="schema_OB11Notify"></a>
<a id="tocSob11notify"></a>
<a id="tocsob11notify"></a>

```json
{
  "request_id": 0,
  "invitor_uin": 0,
  "invitor_nick": "string",
  "group_id": 0,
  "group_name": "string",
  "message": "string",
  "checked": true,
  "actor": 0,
  "requester_nick": "string"
}

```

OneBot 11 通知信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|request_id|number|true|none||请求ID|
|invitor_uin|number|true|none||邀请者QQ|
|invitor_nick|string|true|none||邀请者昵称|
|group_id|number|true|none||群号|
|group_name|string|true|none||群名称|
|message|string|true|none||附言|
|checked|boolean|true|none||是否已处理|
|actor|number|true|none||操作者QQ|
|requester_nick|string|true|none||申请者昵称|

<h2 id="tocS_OB11User">OB11User</h2>

<a id="schemaob11user"></a>
<a id="schema_OB11User"></a>
<a id="tocSob11user"></a>
<a id="tocsob11user"></a>

```json
{
  "birthday_year": 0,
  "birthday_month": 0,
  "birthday_day": 0,
  "phone_num": "string",
  "email": "string",
  "category_id": 0,
  "user_id": 0,
  "nickname": "string",
  "remark": "string",
  "sex": "string",
  "level": 0,
  "age": 0,
  "qid": "string",
  "login_days": 0,
  "categoryName": "string",
  "categoryId": 0
}

```

OneBot 11 用户信息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|birthday_year|number|false|none||出生年份|
|birthday_month|number|false|none||出生月份|
|birthday_day|number|false|none||出生日期|
|phone_num|string|false|none||手机号|
|email|string|false|none||邮箱|
|category_id|number|false|none||分组ID|
|user_id|number|true|none||QQ号|
|nickname|string|true|none||昵称|
|remark|string|false|none||备注|
|sex|string|false|none||性别|
|level|number|false|none||等级|
|age|number|false|none||年龄|
|qid|string|false|none||QID|
|login_days|number|false|none||登录天数|
|categoryName|string|false|none||分组名称|
|categoryId|number|false|none||分组ID|

<h2 id="tocS_OB11LatestMessage">OB11LatestMessage</h2>

<a id="schemaob11latestmessage"></a>
<a id="schema_OB11LatestMessage"></a>
<a id="tocSob11latestmessage"></a>
<a id="tocsob11latestmessage"></a>

```json
{
  "self_id": 0,
  "user_id": 0,
  "time": 0,
  "real_seq": "string",
  "message_type": "string",
  "sender": {
    "user_id": 0,
    "nickname": "string",
    "card": "string",
    "role": "string"
  },
  "raw_message": "string",
  "font": 0,
  "sub_type": "string",
  "message": {},
  "message_format": "string",
  "post_type": "string",
  "group_id": 0,
  "group_name": "string"
}

```

最后一条消息

### 属性

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|self_id|number|true|none||发送者QQ号|
|user_id|number|true|none||接收者QQ号|
|time|number|true|none||时间戳|
|real_seq|string|true|none||消息序号|
|message_type|string|true|none||消息类型|
|sender|object|true|none||none|
|» user_id|number|true|none||用户QQ号|
|» nickname|string|true|none||用户昵称|
|» card|string|false|none||用户名片|
|» role|string|false|none||用户角色|
|raw_message|string|true|none||原始消息|
|font|number|true|none||字体大小|
|sub_type|string|true|none||子类型|
|message|object|true|none||消息内容|
|message_format|string|true|none||消息格式|
|post_type|string|true|none||发布类型|
|group_id|number|true|none||群号|
|group_name|string|true|none||群名称|
