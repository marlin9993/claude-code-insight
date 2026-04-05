import{a as C,D as T,W as A}from"./constants-B6sX234K.js";import{l as h}from"./formatters-ByvtElMG.js";import{y as k,r as m,o as N}from"./index-DmSkzjqr.js";const $="/api",j={async getSession(e){return(await C.get(`${$}/sessions/${e}`)).data},async getSessionMessages(e,t){return(await C.get(`${$}/sessions/${e}/messages`,{params:t})).data}};function b(e){return O(e).length>0}function O(e){if(!e.message||typeof e.message.content!="object")return[];const t=e.message.content;return Array.isArray(t)?t.filter(n=>n.type==="tool_use"||n.type==="tool_result"):[]}function U(e){if(!e.message)return"";const t=e.message.content;return typeof t=="string"?t:Array.isArray(t)?t.filter(r=>r.type==="text"||r.type==="thinking").map(r=>r.type==="text"?r.text:r.type==="thinking"?`💭 ${r.thinking}`:"").join(`
`):JSON.stringify(t)}function H(e){return e.type==="tool_use"?`🔧 调用工具: ${e.name}`:e.type==="tool_result"?"↩️ 工具返回":"🔧 工具"}function K(e){if(e.type==="tool_use"){const t={工具名称:e.name,调用ID:e.id,参数:e.input};return JSON.stringify(t,null,2)}else if(e.type==="tool_result"){if(e.content&&Array.isArray(e.content)){const t=e.content.filter(n=>n.type==="text").map(n=>n.text).join(`

`);return t||JSON.stringify({工具调用ID:e.tool_use_id,结果类型:"结构化数据",内容项数:e.content.length},null,2)}return JSON.stringify(e,null,2)}return JSON.stringify(e,null,2)}function I(e,t){const n=e.type==="user"?"👤 用户":"🤖 助手",r=t?`

*时间: ${new Date(e.timestamp).toLocaleString("zh-CN")}*`:"";let s="";if(e.message){const l=e.message.content;typeof l=="string"?s=l:Array.isArray(l)&&(s=l.filter(o=>o.type==="text"||o.type==="thinking").map(o=>o.type==="text"?o.text||"":o.type==="thinking"?`💭 ${o.thinking||""}`:"").join(`

`))}return`### ${n}${r}

${s}
`}function D(e){if(e.type==="tool_use"){const t=e.name||"Unknown",n=e.id||"",r=JSON.stringify(e.input,null,2);return`
<details>
<summary>🔧 工具调用: ${t}</summary>

**调用 ID**: \`${n}\`

**参数**:
\`\`\`json
${r}
\`\`\`

</details>
`}else if(e.type==="tool_result"){const t=e.tool_use_id||"";let n="";if(e.content&&Array.isArray(e.content)){const r=e.content.filter(s=>s.type==="text").map(s=>"text"in s?s.text:"").filter(Boolean).join(`

`);r?n=r:n=JSON.stringify(e.content,null,2)}else n=JSON.stringify(e,null,2);return`
<details>
<summary>↩️ 工具返回</summary>

**工具调用 ID**: \`${t}\`

**结果**:
\`\`\`
${n}
\`\`\`

</details>
`}return""}function R(e){return`---
title: ${e.title||`会话 ${e.sessionId}`}
session_id: ${e.sessionId}
export_date: ${e.exportDate}
message_count: ${e.messageCount}
${e.project?`project: ${e.project}`:""}
---

`}function J(e,t,n={}){const{includeMetadata:r=!0,includeToolCalls:s=!0,includeTimestamps:l=!0,title:d}=n;let o="";if(r){const c={sessionId:e,exportDate:new Date().toISOString(),messageCount:t.length,title:d};o+=R(c),o+=`
`}return o+=`# ${d||"Claude Code 会话记录"}

`,t.length>5&&(o+=`## 目录

`,t.forEach((c,a)=>{const i=c.type==="user"?"用户":"助手",u=c.message?.content&&typeof c.message.content=="string"?c.message.content.substring(0,T.MD_CONTENT_PREVIEW_LENGTH):"消息";o+=`${a+1}. [${i}: ${u}...](#message-${a+1})
`}),o+=`
`),o+=`## 会话内容

`,t.forEach((c,a)=>{if(o+=`<a id="message-${a+1}"></a>

`,o+=I(c,l),s&&b(c)){const i=O(c);o+=`#### 工具调用/返回

`,i.forEach(u=>{o+=D(u)}),o+=`
`}o+=`---

`}),r&&(o+=`
*本文档由 Claude Code 历史记录查看器自动生成*
`,o+=`*导出时间: ${new Date().toLocaleString("zh-CN")}*
`),o}function P(e,t){const n=new Blob([e],{type:"text/markdown;charset=utf-8"}),r=URL.createObjectURL(n),s=document.createElement("a");s.href=r,s.download=t,document.body.appendChild(s),s.click(),document.body.removeChild(s),URL.revokeObjectURL(r)}const f=m(null),v=m(!1),y=m(!1),p=m(null),E=m(!1),g=new Map;let w=0;function z(){const e=new Map;let t=!1;function n(){const a=window.location.protocol==="https:"?"wss:":"ws:",i=window.location.host;return`${a}//${i}/ws`}function r(){if(f.value?.readyState===WebSocket.OPEN){v.value=!0;return}if(y.value)return;y.value=!0,E.value=!1;const a=n();try{f.value=new WebSocket(a),f.value.onopen=()=>{h.info("WebSocket 连接成功"),v.value=!0,y.value=!1,p.value&&(clearTimeout(p.value),p.value=null)},f.value.onmessage=i=>{try{const u=JSON.parse(i.data),S=g.get(u.type);S&&S.forEach(x=>{try{x(u.data)}catch(_){h.error(`处理消息类型 ${u.type} 时出错:`,_)}})}catch(u){h.error("解析 WebSocket 消息失败:",u)}},f.value.onclose=()=>{h.info("WebSocket 连接断开"),v.value=!1,y.value=!1,f.value=null,!E.value&&!p.value&&(p.value=setTimeout(()=>{p.value=null,r()},A.RECONNECT_DELAY))},f.value.onerror=i=>{h.error("WebSocket 错误:",i),y.value=!1}}catch(i){h.error("创建 WebSocket 连接失败:",i),y.value=!1}}function s(){t||(w++,t=!0),r()}function l(a,i){g.has(a)||g.set(a,new Set),g.get(a).add(i),e.has(a)||e.set(a,new Set),e.get(a).add(i)}function d(a,i){const u=g.get(a);u&&(u.delete(i),u.size===0&&g.delete(a));const S=e.get(a);S&&(S.delete(i),S.size===0&&e.delete(a))}function o(){e.forEach((a,i)=>{a.forEach(u=>{d(i,u)})}),e.clear()}function c(){t&&(w--,t=!1),!(w>0)&&(w=0,E.value=!0,p.value&&(clearTimeout(p.value),p.value=null),f.value&&(f.value.close(),f.value=null),v.value=!1,y.value=!1,g.clear())}return k(()=>{o(),c()}),{connected:v,connecting:y,connect:s,disconnect:c,on:l,off:d}}function B(e){function t(n){const r=n.target;if(r.tagName==="INPUT"||r.tagName==="TEXTAREA"||r.contentEditable==="true"||r.closest(".el-input")||r.closest(".el-textarea"))return;const s=[];(n.ctrlKey||n.metaKey)&&s.push(n.ctrlKey?"Ctrl":"Cmd"),n.altKey&&s.push("Alt"),n.shiftKey&&s.push("Shift");let l=n.key;l===" "&&(l="Space"),l==="Escape"&&(l="Escape"),l==="Enter"&&(l="Enter");const d=s.length>0?[...s,l].join("+"):l;e[d]&&(n.preventDefault(),n.stopPropagation(),e[d]())}N(()=>{window.addEventListener("keydown",t)}),k(()=>{window.removeEventListener("keydown",t)})}const G={SEARCH:"Ctrl+K",REFRESH:"Ctrl+R",GOTO_HOME:"Ctrl+H",GOTO_STATS:"Ctrl+S",GOTO_STARRED:"Ctrl+Shift+S",CLOSE:"Escape",EXPORT:"Ctrl+E"};export{G as S,B as a,K as b,H as c,P as d,J as e,U as f,O as g,b as h,j as s,z as u};
