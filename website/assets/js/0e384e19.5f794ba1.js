"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[671],{3905:function(e,t,n){n.d(t,{Zo:function(){return u},kt:function(){return f}});var r=n(7294);function o(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function i(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function a(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?i(Object(n),!0).forEach((function(t){o(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):i(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,r,o=function(e,t){if(null==e)return{};var n,r,o={},i=Object.keys(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||(o[n]=e[n]);return o}(e,t);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(e);for(r=0;r<i.length;r++)n=i[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(o[n]=e[n])}return o}var s=r.createContext({}),c=function(e){var t=r.useContext(s),n=t;return e&&(n="function"==typeof e?e(t):a(a({},t),e)),n},u=function(e){var t=c(e.components);return r.createElement(s.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},d=r.forwardRef((function(e,t){var n=e.components,o=e.mdxType,i=e.originalType,s=e.parentName,u=l(e,["components","mdxType","originalType","parentName"]),d=c(n),f=o,m=d["".concat(s,".").concat(f)]||d[f]||p[f]||i;return n?r.createElement(m,a(a({ref:t},u),{},{components:n})):r.createElement(m,a({ref:t},u))}));function f(e,t){var n=arguments,o=t&&t.mdxType;if("string"==typeof e||o){var i=n.length,a=new Array(i);a[0]=d;var l={};for(var s in t)hasOwnProperty.call(t,s)&&(l[s]=t[s]);l.originalType=e,l.mdxType="string"==typeof e?e:o,a[1]=l;for(var c=2;c<i;c++)a[c]=n[c];return r.createElement.apply(null,a)}return r.createElement.apply(null,n)}d.displayName="MDXCreateElement"},9881:function(e,t,n){n.r(t),n.d(t,{frontMatter:function(){return l},contentTitle:function(){return s},metadata:function(){return c},toc:function(){return u},default:function(){return d}});var r=n(7462),o=n(3366),i=(n(7294),n(3905)),a=["components"],l={id:"intro",title:"Introduction",sidebar_label:"Introduction",slug:"/"},s=void 0,c={unversionedId:"intro",id:"intro",isDocsHomePage:!1,title:"Introduction",description:"Goals",source:"@site/docs/intro.md",sourceDirName:".",slug:"/",permalink:"/gatekeeper/website/docs/",editUrl:"https://github.com/open-policy-agent/gatekeeper/edit/master/website/docs/intro.md",tags:[],version:"current",frontMatter:{id:"intro",title:"Introduction",sidebar_label:"Introduction",slug:"/"},sidebar:"docs",next:{title:"Installation",permalink:"/gatekeeper/website/docs/install"}},u=[{value:"Goals",id:"goals",children:[]},{value:"How is Gatekeeper different from OPA?",id:"how-is-gatekeeper-different-from-opa",children:[{value:"Admission Webhook Fail-Open by Default",id:"admission-webhook-fail-open-by-default",children:[]}]}],p={toc:u};function d(e){var t=e.components,n=(0,o.Z)(e,a);return(0,i.kt)("wrapper",(0,r.Z)({},p,n,{components:t,mdxType:"MDXLayout"}),(0,i.kt)("h2",{id:"goals"},"Goals"),(0,i.kt)("p",null,"Every organization has policies. Some are essential to meet governance and legal requirements. Others help ensure adherence to best practices and institutional conventions. Attempting to ensure compliance manually would be error-prone and frustrating. Automating policy enforcement ensures consistency, lowers development latency through immediate feedback, and helps with agility by allowing developers to operate independently without sacrificing compliance."),(0,i.kt)("p",null,"Kubernetes allows decoupling policy decisions from the inner workings of the API Server by means of ",(0,i.kt)("a",{parentName:"p",href:"https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/"},"admission controller webhooks"),", which are executed whenever a resource is created, updated or deleted. Gatekeeper is a validating (mutating TBA) webhook that enforces CRD-based policies executed by ",(0,i.kt)("a",{parentName:"p",href:"https://github.com/open-policy-agent/opa"},"Open Policy Agent"),", a policy engine for Cloud Native environments hosted by CNCF as an incubation-level project."),(0,i.kt)("p",null,"In addition to the ",(0,i.kt)("inlineCode",{parentName:"p"},"admission")," scenario, Gatekeeper's audit functionality allows administrators to see what resources are currently violating any given policy."),(0,i.kt)("p",null,"Finally, Gatekeeper's engine is designed to be portable, allowing administrators to detect and reject non-compliant commits to an infrastructure-as-code system's source-of-truth, further strengthening compliance efforts and preventing bad state from slowing down the organization."),(0,i.kt)("h2",{id:"how-is-gatekeeper-different-from-opa"},"How is Gatekeeper different from OPA?"),(0,i.kt)("p",null,"Compared to using ",(0,i.kt)("a",{parentName:"p",href:"https://www.openpolicyagent.org/docs/kubernetes-admission-control.html"},"OPA with its sidecar kube-mgmt")," (aka Gatekeeper v1.0), Gatekeeper introduces the following functionality:"),(0,i.kt)("ul",null,(0,i.kt)("li",{parentName:"ul"},"An extensible, parameterized policy library"),(0,i.kt)("li",{parentName:"ul"},'Native Kubernetes CRDs for instantiating the policy library (aka "constraints")'),(0,i.kt)("li",{parentName:"ul"},'Native Kubernetes CRDs for extending the policy library (aka "constraint templates")'),(0,i.kt)("li",{parentName:"ul"},"Audit functionality")),(0,i.kt)("h3",{id:"admission-webhook-fail-open-by-default"},"Admission Webhook Fail-Open by Default"),(0,i.kt)("p",null,"Currently Gatekeeper is defaulting to using ",(0,i.kt)("inlineCode",{parentName:"p"},"failurePolicy\u200b: \u200bIgnore")," for admission request webhook errors. The impact of\nthis is that when the webhook is down, or otherwise unreachable, constraints will not be\nenforced. Audit is expected to pick up any slack in enforcement by highlighting invalid\nresources that made it into the cluster."),(0,i.kt)("p",null,"If you would like to switch to fail closed, please see our ",(0,i.kt)("a",{parentName:"p",href:"/gatekeeper/website/docs/failing-closed"},"documentation")," on how to do so and some things you should consider before doing so."))}d.isMDXComponent=!0}}]);