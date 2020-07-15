## 内部机制

### CGO 生成的中间文件

在构建一个 cgo 包时增加一个 `-work` 输出中间生成所在目录并且在构建完成时保留中间文件.

对于比较简单的 cgo 代码可以直接手工调用 `go tool cgo` 命令来查看生成的中间文件.

在一个 Go 源文件当中, 如果出现 `import "C"` 指令则表示将调用 cgo 命令生成对应的中间文件. 下面是生成的中间文件的简单
示意图:


包含有 4 个 Go 文件, 其中 nocgo 开头的文件中没有 `import "C"` 指令, 其他的 2 个文件则包含了 cgo 代码. cgo 命令会为每个包含 cgo 代码的 Go 文件创建 2 个中间文件, 比如 main.go 会分别创建 **main.cgo1.go** 和 **main.cgo2.c** 两个中间文件, cgo当中是 Go 代码部分和 C 代码部分.

然后会为整个包创建一个 **_cgo_gotypes.go** Go 文件, 其中包含 Go 语言部分辅助代码. 此外还创建一个 **_cgo_export.h** 和 **_cgo_export.c** 文件, 对应 Go 语言导出到 C 语言的类型和函数.

```cgo
package main

/*
int sum(int a, int b) { 
  return a+b; 
}
*/
import "C"

func main() {
    println(C.sum(1, 1))
}
```

使用 `go tool cgo main.go` 生成的中间文件如下:

```
_cgo_export.c
_cgo_export.h
_cgo_flags
_cgo_gotypes.go
_cgo_main.c
_cgo_.o
main.cgo1.go
main.cgo2.c
```

其中 **main.cgo1.go [1]** 的代码如下:

```
package main

/*
int sum(int a, int b) {
  return a+b; 
}
*/
import _ "unsafe"

func main() {
	println((_Cfunc_sum)(1, 1))
}
```

其中 `C.sum(1,1)` 函数调用替换成了 `(_Cfunc_sum)(1,1)`. 每一个 `C.xxx` 形式的函数都会被替换成 `_Cfunc_xxx` 格式的纯 Go 函数, 其中前缀 `_Cfunc_` 表示这是一个C函数, 对应一个私有的 Go 桥接函数.

**_Cfunc_sum** 函数在 cgo 生成的 **_cgo_gotypes.go [2]** 文件当中定义:

```
//go:cgo_import_static _cgo_7b5139e7c7da_Cfunc_sum
//go:linkname __cgofn__cgo_7b5139e7c7da_Cfunc_sum _cgo_7b5139e7c7da_Cfunc_sum
var __cgofn__cgo_7b5139e7c7da_Cfunc_sum byte
var _cgo_7b5139e7c7da_Cfunc_sum = unsafe.Pointer(&__cgofn__cgo_7b5139e7c7da_Cfunc_sum)


//go:cgo_unsafe_args
func _Cfunc_sum(p0 _Ctype_int, p1 _Ctype_int) (r1 _Ctype_int) {
	_cgo_runtime_cgocall(_cgo_7b5139e7c7da_Cfunc_sum, uintptr(unsafe.Pointer(&p0)))
	if _Cgo_always_false {
		_Cgo_use(p0)
		_Cgo_use(p1)
	}
	return
}
```

`_Cfunc_sum` 函数的参数和返回值 `_Ctype_int` 类型对应 `C.int` 类型, 命名的规则和 `_C_func_xxx` 类似, 
不同的前缀用于区分函数和类型.

`_cgo_runtime_cgocall` 对应 `runtime.cgocall` 函数, 声明如下:

```
func runtime.cgocall(fn, arg unsafe.Pointer) int32
```

> 第一个参数是 C 语言函数的地址
> 第二个参数是存储 C 语言函数对应的参数结构体(参数和返回值)的地址

在此例当中, 被传入C语言函数 **_cgo_7b5139e7c7da_Cfunc_sum** 也是 cgo 生成的中间函数. 函数定义在**main.cgo2.c [3]** 当中.

```
void _cgo_7b5139e7c7da_Cfunc_sum(void *v)
{
	struct {
		int p0;
		int p1;
		int r;
		char __pad12[4];
	} __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
	char *_cgo_stktop = _cgo_topofstack();
	__typeof__(_cgo_a->r) _cgo_r;
	_cgo_tsan_acquire();
	_cgo_r = sum(_cgo_a->p0, _cgo_a->p1);
	_cgo_tsan_release();
	_cgo_a = (void*)((char*)_cgo_a + (_cgo_topofstack() - _cgo_stktop));
	_cgo_a->r = _cgo_r;
	_cgo_msan_write(&_cgo_a->r, sizeof(_cgo_a->r));
}
```

此函数参数只有一个 `void*` 的指针, 函数没有返回值. 真实的 sum 函数的函数参数和返回值通过唯一的参数指针类实现.

```
struct {
		int p0;
		int p1;
		int r;
		char __pad12[4];
} __attribute__((__packed__, __gcc_struct__)) *_cgo_a = v;
```

其中, p0, p1分别对应 sum 的第一个和第二个参数, r 对应 sum 的返回值. `_pad12` 用于填充结构体保证对齐CPU机器字的整数倍.

> 然后从参数执行的结构体获取参数后开始调用真实的C语言版sum函数, 并且将返回值保存到结构体的返回值对应的成员.


因为 Go 语言和 C 语言有着不同的内存模型和函数调用规范, 其中 `_cgo_topofstack` 函数相关的代码用于 C 函数调用后恢复调用栈. `_cgo_tsan_acquire` 和 `_cgo_tsan_release` 则是用于扫描 CGO 相关函数的指针总相关检查.

调用链:

文件链:

```
main.go -> main.cgo1.go -> _cgo_gotypes.go -> main.cgo2.c
```