# ESI Language Specification 1.0

**W3C Note 04 August 2001**

- **This version:** http://www.w3.org/TR/2001/NOTE-esi-lang-20010804
- **Latest version:** http://www.w3.org/TR/esi-lang

## Authors
- Mark Tsimelzon, Akamai Technologies
- Bill Weihl, Akamai Technologies
- Joseph Chung, Art Technology Group
- Dan Frantz, BEA Systems
- John Basso, Circadence Corporation
- Chris Newton, Digital Island, Inc.
- Mark Hale, Interwoven, Inc.
- Larry Jacobs, Oracle Corporation
- Conleth O'Connell, Vignette Corporation

**Editor:** Mark Nottingham, Akamai Technologies

## Abstract

This specification defines ESI 1.0, the Edge Side Includes language, which allows content assembly by HTTP surrogates, by providing an in-markup XML-based language.

## 1. Introduction

Edge Side Includes (ESI) is an XML-based markup language that provides a means to assemble resources in HTTP clients. Unlike other in-markup languages, ESI is designed to leverage client tools like caches to improve end-user perceived performance, reduce processing overhead on the origin server, and enhanced availability. ESI allows for dynamic content assembly at the edge of the network, whether it is in a Content Delivery Network, end-user's browser, or in a "Reverse Proxy" right next to the origin server.

ESI is primarily intended for processing on surrogates (intermediaries that operate on behalf of the origin server, also known as "Reverse Proxies") that understand the ESI language. However, its application is not restricted to these devices. The control of where ESI is processed is addressed in the Edge Architecture Specification. Its capability token is:

```
ESI/1.0
```

ESI allows surrogates to treat parts of pages as cacheable resources, which gives them the ability to serve resources from cache in more situations.

### 1.1 Relationship to Other Standards

ESI is an XML language designed to be interposed into markup to provide logic and dispatch services, targeted for processing after the markup has left the origin server, but before it is paginated by the end user's client. As a result, the markup that is emitted by the origin server is not valid; it contains interposed elements from the ESI namespace. Additionally, it may contain native markup (e.g., HTML) that will not exist in the paginated entity.

XInclude is a W3C effort to standardize a general inclusion mechanism for XML. The inclusion aspect of ESI is somewhat similar to XInclude, with additional semantics for failure handling. Additionally, ESI processing is targeted (usually, to surrogates), while XInclude doesn't define explicit or implicit targeting of processing.

Additionally, ESI borrows some portions of the XSLT language for logic and processing control.

### 1.2 Relationship to in-Markup Content Generation and Client Scripting Languages

There are several proprietary in-markup languages available (e.g., PHP, SSI, etc.), as well as a few standardized solutions (e.g., ECMAScript, etc.). ESI is not intended to replace these languages. It is expected that ESI will be used in concert with both content generation and client scripting markup.

When ESI is processed on the same device as other markup (e.g., Server Side Includes), some form of precedence in processing will need to be defined. This determination is currently implementation-specific.

## 2. ESI Functional Overview

The ESI language is conceptually similar in many ways to the Server Side Includes (SSI) function found in many web servers. It is an in-markup scripting language that is interpreted before the page is served to the client.

Version 1.0 includes the following functionality:

- **Inclusion** - ESI can compose pages by assembling included content, which is fetched from the network. This allows each such *fragment* to have its own metadata (e.g., cacheability and handling information) separately associated.
- **Variable support** - ESI 1.0 supports the use of variables based on HTTP request attributes in a manner reminiscent of the Common Gateway Interface. These variables can be used by ESI statements or written directly into the processed markup.
- **Conditional processing** - ESI allows conditional logic with Boolean comparisons to be used to influence how a template is processed.
- **Exception and error handling** - ESI provides for specification of alternate and default resources in a number of situations.

The ESI assembly model is comprised of a *template* containing *fragments*. The template is the container for assembly, with instructions for the retrieval of fragments, and is the resource associated with the URL the end user requests. It includes ESI elements that instruct *ESI Processors* (clients that understand ESI) to fetch and include a fragment's URI. The fragments themselves can be any textual Web resource, typically HTML markup.

Because fragments are separate resources, they can be assigned their own cacheability and handling information. For example, a cache time-to-live (TTL) of several days could be appropriate for the template, but a fragment containing a frequently-changing story or ad may require a much lower TTL. Some fragments may require being marked uncacheable.

## 3. ESI Elements

ESI elements are XML, in an ESI-specific XML Namespace. This allows them to be embedded in many common Web document formats, including HTML and XML-based server-side processing languages. ESI Processors parse but do not process elements outside of the ESI namespace. When an ESI Processor processes a template, ESI elements are stripped from the output.

The XML Namespace for ESI 1.0 is:

```
http://www.edge-delivery.org/esi/1.0
```

Future versions of and extensions to ESI will use distinct namespaces. Typically, documents will declare the ESI namespace in the top-level element; for templates, this would be the `<html>` tag, while in fragments this could be a `<div>` tag wrapping the entire fragment.

ESI element names and attribute names are always lowercase.

### 3.1 include

The include element specifies a fragment for assembly and allows for two optionally specified behaviors. include is an empty element; it does not have a closing tag.

```xml
<esi:include src="URI" alt="URI" onerror="continue" />
```

For example:

```xml
<esi:include src="http://example.com/1.html" alt="http://bak.example.com/2.html" onerror="continue"/>
<esi:include src="http://example.com/search?query=$(QUERY_STRING{query})"/>
```

The include statement tells ESI Processors to fetch the resource specified by the src attribute. This can be a simple URI, as shown in the first example, or can include variables (see "Variables"), as shown in the second. In either case, the final attribute value must be a valid URI. Relative URIs will be resolved relative to the template. The resulting object will replace the element in the markup served to the client.

ESI Processor implementations may limit the number of includes used in a single ESI resource. Additionally, they may limit the number and/or depth of included documents that will be recursed. This assures that ESI processing does not monopolize resources or impact end-user perceived performance.

The optional `alt` attribute specifies an alternative resource if the src is not found. The requirements for the value are the same as those for src. Some ESI Processors may not implement this attribute, depending on its applicability; for example, surrogates near the origin server typically cannot usefully process them.

If an ESI Processor can fetch neither the src nor the alt, it returns a HTTP status code greater than 400 with an error message, unless the `onerror` attribute is present. If it is, and `onerror="continue"` is specified, ESI Processors will delete the include element silently.

### 3.2 inline

ESI fragments need not be fetched independently by the ESI processor. The inline element provides a way to demarcate fragments, embedded in the HTTP response. These fragments are stored and assembled in the ESI processor as independently included fragments are handled. Inline has a closing tag.

```xml
<esi:inline name="URI" fetchable="{yes | no}"> 
    fragment to be stored within an ESI processor 
</esi:inline>
```

The inline statement is used to demarcate ESI fragments. The fragment is embedded within an HTTP response to an ESI processor. The ESI processor will parse response and extract all inline fragments and store them independently, under the URI specified.

Some inline fragments are only delivered as part of an HTTP response for another object. These are said to be not independently fetchable by the ESI processor. When a non fetchable fragment is needed by the ESI processor, the ESI processor must request the object from which the inline fragment was extracted.

An independently fetchable fragment may be requested by the ESI processor by using its name as the URI.

Implementation of inline is optional; ESI Processors use the capability token:

```
ESI-Inline/1.0
```

to advertise their willingness to process this tag.

### 3.3 choose | when | otherwise

These conditional elements add the ability to perform logic based on expressions. All three must have an end tag.

```xml
<esi:choose> 
    <esi:when test="...">
        ...
    </esi:when> 
    <esi:when test="...">
        ...
    </esi:when>
    <esi:otherwise> 
        ...
    </esi:otherwise>
</esi:choose>
```

Every choose element must contain at least one when element, and may optionally contain exactly one otherwise element. No other ESI elements or non-ESI markup can be direct children of a choose element.

ESI processors will execute the first when statement whose test attribute evaluates truthfully, and then exit the choose element. If no when element evaluates to true, and an otherwise element is present, that element's content will be executed. See "ESI Expressions" for the syntax of the test attribute's value.

ESI elements as well as non-ESI markup can be included inside when or otherwise elements.

For example:

```xml
<esi:choose> 
    <esi:when test="$(HTTP_COOKIE{group})=='Advanced'"> 
        <esi:include src="http://www.example.com/advanced.html"/> 
    </esi:when> 
    <esi:when test="$(HTTP_COOKIE{group})=='Basic User'">
        <esi:include src="http://www.example.com/basic.html"/>
    </esi:when> 
    <esi:otherwise> 
        <esi:include src="http://www.example.com/new_user.html"/> 
    </esi:otherwise>
</esi:choose>
```

### 3.4 try | attempt | except

Exception handling is provided by the try element, which must contain exactly one instance of both an attempt and an except element (all with end tags):

```xml
<esi:try>
    <esi:attempt> 
        ...
    </esi:attempt> 
    <esi:except> 
        ...
    </esi:except>
</esi:try>
```

Valid children of try are attempt and except; no other ESI or non-ESI markup can be a direct child. attempt and except may contain valid ESI or non-ESI markup.

ESI Processors first execute the contents of attempt. A failed ESI include statement will trigger an error and cause the ESI Processor to execute the contents of the except element. Statements other than include and inline will not trigger this error.

In this example, the attempt is to fetch an ad. If the ad fetch fails, a static link will be included instead.

```xml
<esi:try> 
    <esi:attempt>
        <esi:comment text="Include an ad"/> 
        <esi:include src="http://www.example.com/ad1.html"/> 
    </esi:attempt>
    <esi:except> 
        <esi:comment text="Just write some HTML instead"/> 
        <a href="www.akamai.com">www.example.com</a>
    </esi:except> 
</esi:try>
```

### 3.5 comment

The comment element allows developers to comment their ESI instructions, without making the comments available in the processor's output. comment is an empty element, and must not have an end tag.

```xml
<esi:comment text="..." />
```

For example:

```xml
<esi:comment text="the following animation will have a 24 hr TTL." />
```

Comments are not evaluated by ESI Processors; they are deleted before output. Comments that need to be visible in the output should use standard XML/HTML comment syntax.

### 3.6 remove

The remove element allows for specification of non-ESI markup for output if ESI processing is not enabled.

```xml
<esi:remove> ... </esi:remove>
```

For example:

```xml
<esi:include src="http://www.example.com/ad.html"/> 
<esi:remove> 
  <a href="http://www.example.com">www.example.com</a>
</esi:remove>
```

Normally, when this block is processed, the ESI Processor fetches the ad.html resource and includes it in the template while silently discarding the remove element and its contents.

If for some reason ESI processing is not enabled, all of the elements will be passed through to clients, which will ignore markup it doesn't understand.

With Web clients, this works because browsers ignore invalid HTML, such as `<esi:...>` and `</esi:...>` elements, leaving the HTML a element and its content.

The remove statement cannot include nested ESI markup.

### 3.7 vars

To include an ESI variable in markup outside an ESI block, use the vars element.

```xml
<esi:vars> ... </esi:vars>
```

For example:

```xml
<esi:vars>
  <img src="http://www.example.com/$(HTTP_COOKIE{type})/hello.gif"/>
</esi:vars>
```

See "ESI Variables" for more information about variables.

### 3.8 <!--esi ...-->

This is a special construct to allow HTML marked up with ESI to render without processing. ESI Processors will remove the start ("<!--esi") and end ("-->") when the page is processed, while still processing the contents. If the page is not processed, it will remain, becoming an HTML/XML comment tag. For example:

```html
<!--esi  
  <p><esi:vars>Hello, $(HTTP_COOKIE{name})!</esi:vars></p>
-->
```

This assures that the ESI markup will not interfere with the rendering of the final HTML if not processed.

## 4. ESI Variables

ESI 1.0 supports the following read-only variables, which are based on the client's HTTP request line and headers:

| Variable Name          | HTTP Header     | Substructure Type    | Example                  |
| ---------------------- | --------------- | -------------------- | ------------------------ |
| HTTP_ACCEPT_LANGUAGE   | Accept-Language | list                 | da, en-gb, en            |
| HTTP_COOKIE            | Cookie          | dictionary           | id=571; visits=42        |
| HTTP_HOST              | Host            | -                    | esi.xyz.com              |
| HTTP_REFERER           | Referer         | -                    | http://roberts.xyz.com/  |
| HTTP_USER_AGENT        | User-Agent      | dictionary (special) | Mozilla; MSIE 5.5        |
| QUERY_STRING           | -               | dictionary           | first=Robin&last=Roberts |

Variable names are always uppercase. To reference a variable, surround the name with parenthesis and append a dollar sign ($).

For example:

```
$(HTTP_HOST)
```

### 4.1 Variable Substructure Access

By default, ESI variables are evaluated in a string context. However, some which represent more complex data will make automatically parsed and typed data available.

To access a variable's substructure, the variable name should be appended with braces containing the key which is being accessed. For example:

```
$(HTTP_COOKIE{username})
```

Variables capable of subaccess can be classified as dictionaries or lists, as outlined in the "ESI Variables" table.

#### 4.1.1 Dictionary

Variables identified as dictionaries make access to strings available through their appropriate keys. Dictionary keys are case-sensitive. For example:

```
$(HTTP_COOKIE{visits})
```

evaluates to "42" if the Cookie header contains the string "visits=42".

The dictionary associated with the User-Agent header contains only three special values; browser, version and os.

- `browser` allows access to the ESI Processor's determination of the User-Agent's browser, and may be either "MOZILLA", "MSIE", or "OTHER".
- `version` allows access to the User-Agent's browser version.
- `os` allows access to the Processor's determination of the User-Agent's platform, and may evaluate to either "WIN", "MAC", "UNIX" or "OTHER".

#### 4.1.2 List

Variables identified as lists will return a boolean value depending on whether the requested value is present. For example:

```
$(HTTP_ACCEPT_LANGUAGE{en-gb})
```

will return true if the string "en-gb" is in the appropriate header; otherwise, it will return false.

### 4.2 Variable Default Values

Variables whose values are empty, nonexistent variables and undefined substructures of variables will evaluate to an empty string when they are accessed. The logical or operator can be used to specify a default value in an expression where desirable, in the form:

```
$(VARIABLE|default)
```

For example:

```xml
<esi:include src="http://example.com/$(HTTP_COOKIE{id}|default).html"/>
```

will result in the ESI Processor fetching the following URI if the cookie "id" is not in the request:

```
http://example.com/default.html
```

As with other literals, if whitespace needs to be specified, the default value must be single-quoted:

```
$(HTTP_COOKIE{first_name}|'new user')
```

## 5. ESI Expressions

Conditional elements use expressions (in their test attributes) to determine how to apply the contained elements. Expressions consist of operators, variables and literals, and evaluate to true or false.

Single quotes are used within an expression to delimit literals. For example:

```
$(HTTP_HOST) == 'example.com'
```

Whitespace is optional around all operators, except `has`, which must be surrounded by whitespace.

### 5.1 Operators

The following set of unary and binary logical operators are supported by ESI expressions, listed in order of decreasing precedence:

| Operator              | Type           |
| --------------------- | -------------- |
| ==, !=, <, >, <=, >= | comparison     |
| !                     | unary negation |
| &                     | logical and    |
| \|                    | logical or     |

Operands associate from left to right. Sub-expressions can be grouped with parentheses in order to explicitly specify association.

If both operands are numeric, the expression is evaluated numerically. If either binary operand is non-numeric, both operands are evaluated as strings. After expansion, variables are loosely typed, but care should be taken. For example, a version reported as 3.01.23 or 1.05a will not test as a number.

The behavior of comparisons which incompatibly typed operators is undefined. If an operand is empty or undefined, the expression will always evaluate to false.

Logical operators ("&", "|", "!") can be used to qualify expressions, but cannot be used as comparators themselves.

For example, the following expressions are correct:

```
!(1==1)
!('a'<='c')
(1==1)|('abc'=='def')
(4!=5)&(4==5)
```

while these will not evaluate:

```
(1 & 4)
("abc" | "edf")
```

## 6. Protocol Considerations

When an ESI template is processed, a separate request will need to be made for each include encountered. Implementations may use the original request's headers (e.g., Cookie, User-Agent, etc.) when doing so. Additionally, response headers from fragments (e.g., Set-Cookie, Server, Cache-Control, Last-Modified) may be ignored, and should not influence the assembled page.

## Appendix A: Acknowledgements

The Authors would like to thank the following for the early and continuing development of the ESI language: Sriram Sankar, Andy Davis, Sam Gendler, Joshua Silver, Jay Parikh, Inbar Zamir, Ziv Katzir, and Chris Weikart. 