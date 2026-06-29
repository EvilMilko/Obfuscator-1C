<h1 align="center">Обфускатор 1С (BSL) — расширенный форк</h1>

<p align="center">
  Форк <a href="https://github.com/LazarenkoA/Obfuscator-1C">LazarenkoA/Obfuscator-1C</a> с готовым CLI, исправлениями ошибок компиляции и новыми возможностями.
</p>

<p align="center">
  <a href="https://github.com/EvilMilko/Obfuscator-1C/releases/latest">
    <img alt="Скачать" src="https://img.shields.io/badge/ скачать-v1.0.0-blue">
  </a>
  <img alt="Язык" src="https://img.shields.io/badge/язык-Go-00ADD8">
  <img alt="Платформа" src="https://img.shields.io/badge/платформа-1С-yellow">
  <img alt="Лицензия" src="https://img.shields.io/badge/license-MIT-green">
</p>

---

## О проекте

Инструмент для обфускации исходного кода на языке **1С:Предприятие (BSL)**. Усложняет реверс-инжиниринг: шифрует строки, запутывает условия, заменяет циклы на `Перейти`, прячет вызовы за `Выполнить()/Вычислить()` и генерирует мусорный код.

Работает на базе [1c-language-parser](https://github.com/LazarenkoA/1c-language-parser).

### Что добавлено относительно оригинала

| Возможность | Оригинал | Этот форк |
|---|---|---|
| **CLI-инструмент** | Только Go API | Готовый `1c-obfuscator.exe` |
| **Пакетная обработка** | Нет | Рекурсивный обход каталога выгрузки |
| **JSON-конфигурация** | Нет | `config.json` |
| **BOM / пустые файлы / panic** | Ошибки | Корректная обработка, копирование оригинала при сбое |
| **Dangling goto labels** | Ошибка компиляции | `ensureLabelFollowedByStatement()` — авто-фикс |
| **Вложенные циклы** | Заменяет Прервать/Продолжить из внешнего цикла | `replaceBreakContinue()` — не заходит во вложенные |
| **RepExpByEval** | Прямая замена на Выполнить/Вычислить | Паттерн `_ТранзакцияАктивна()` |
| **NoEvalFuncs** | Нет | Исключение функций из Eval-обфускации |

---

## Быстрый старт

### Скачать готовый exe

[**Releases →**](https://github.com/EvilMilko/Obfuscator-1C/releases/latest)

### Использование

```bash
1c-obfuscator.exe -dir="D:\MyConfig"
```

Утилита создаст копию каталога `D:\MyConfig_obfuscated` со всеми обфусцированными `.bsl` файлами. Остальные файлы копируются без изменений.

```
D:\MyConfig\                          D:\MyConfig_obfuscated\
├── Catalogs\                         ├── Catalogs\
│   ├── Товары\                       │   ├── Товары\
│   │   └── Ext\ObjectModule.bsl  ──► │   │   └── Ext\ObjectModule.bsl  (обфусцирован)
│   └── ...                           │   └── ...
├── CommonModules\                    ├── CommonModules\
│   └── ...                           │   └── ...                      (обфусцированы)
├── Configuration.xml             ──► ├── Configuration.xml            (копия)
└── ...                               └── ...
```

### config.json

```json
{
  "RepExpByTernary": true,
  "RepLoopByGoto": true,
  "RepExpByEval": true,
  "HideString": true,
  "ChangeConditions": true,
  "AppendGarbage": true,
  "CallStackHell": true,
  "LineBreaks": false,
  "NoEvalFuncs": []
}
```

| Параметр | Описание |
|---|---|
| `RepExpByTernary` | Заменять простые выражения тернарными операторами `?(...)` |
| `RepLoopByGoto` | Заменять циклы `Пока`/`Для` на безусловные переходы `Перейти` |
| `RepExpByEval` | Прятать вызовы методов в `Выполнить()` / `Вычислить()` |
| `HideString` | Шифровать строковые константы (Base64 + XOR) |
| `ChangeConditions` | Запутывать логические условия добавлением фиктивных `И` |
| `AppendGarbage` | Добавлять неиспользуемый мусорный код |
| `CallStackHell` | Прятать выражения за цепочками фейковых функций |
| `LineBreaks` | `true` — сохранять переносы строк, `false` — всё в одну строку |
| `NoEvalFuncs` | Список имён функций, в которых **не** применять `RepExpByEval` |

---

## Использование как Go-библиотеки

```go
package main

import (
    "context"
    "fmt"

    "github.com/EvilMilko/Obfuscator-1C/obfuscator"
)

func main() {
    code := `Процедура Тест()
        Сообщить("Привет");
    КонецПроцедуры`

    obf := obfuscator.NewObfuscatory(context.Background(), obfuscator.Config{
        RepExpByTernary:  true,
        RepLoopByGoto:    true,
        RepExpByEval:     true,
        HideString:       true,
        ChangeConditions: true,
        AppendGarbage:    true,
        CallStackHell:    true,
        NoEvalFuncs:      []string{"_ТранзакцияАктивна"},
    })

    obfuscated, err := obf.Obfuscate(code)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(obfuscated)
}
```

---

## Пример обфускации

**Исходный код:**
```bsl
Процедура Команда1НаСервере()
    Запрос = Новый Запрос;
    Запрос.Текст =
        "ВЫБРАТЬ
        |   Оборудование.Ссылка КАК Ссылка
        |ИЗ
        |   Справочник.Оборудование КАК Оборудование";

    РезультатЗапроса = Запрос.Выполнить();
    ВыборкаДетальныеЗаписи = РезультатЗапроса.Выбрать();

    Пока ВыборкаДетальныеЗаписи.Следующий() Цикл
        Сообщить(ВыборкаДетальныеЗаписи.Ссылка);
    КонецЦикла;
КонецПроцедуры
```

**После обфускации (максимальные настройки):**
```bsl
Процедура Команда1НаСервере() Запрос = Вычислить(raeyrчфлинzpeолюучzт("0LzQn9CT0arQmAHQ...", ?(81 + 94 - 12 / 20 + 70 / 68 < 80 - 21 + 54 - 30, 674, ?(76 / 98 * 9 + 13 + 99 / 57 > 56 * 30 + 81, 727, ?(71 + 97 - 51 > 71 * 63 / 26 * 56 / 93, 33, 666)))));~аqс:Если Не ВыборкаДетальныеЗаписи.Следующий() Тогда Перейти ~sъjm;КонецЕсли;Сообщить(ВыборкаДетальныеЗаписи.Ссылка);Перейти ~аqс;~sъjm:iзylшйбилnowм = ?(88 + 44 > 76 - 93 / 47 + 11, "islmл", ?(59 / 58 < 42 / 71, "roтrзwlv", "vгаzэop"));КонецПроцедуры
```

Кода становится на порядок больше, однако добавленный код — мусорный, он никогда не выполняется и не влияет на производительность.

### Замер производительности

**До обфускации:** ![](img/ShareX_78xr65aMET.png)

**После (максимальные настройки):** ![](img/1cv8_tuEuQCwfJS.png)

---

## Исправления ошибок

### Dangling goto labels (висячие метки)

Оригинальный обфускатор мог сгенерировать метку `Перейти` (`~name:`), за которой следует закрывающий оператор (`КонецЕсли`, `Иначе`, `КонецЦикла`) или другая метка. Это вызывало ошибку компиляции 1С: *«Обнаружено логическое завершение исходного текста модуля»*.

**Решение:** метод `ensureLabelFollowedByStatement()` рекурсивно обходит AST и вставляет no-op оператор (`rndvar = 0`) после каждой висячей/смежной метки **перед** выводом результата. Семантика не меняется — no-op добавляется после всех трансформаций.

### Прервать / Продолжить во вложенных циклах

Оригинал использовал `ast.StatementWalk` для замены `Прервать`/`Продолжить`, который заходил во **вложенные циклы** и заменял их операторы метками **внешнего** цикла.

**Решение:** собственный метод `replaceBreakContinue()` с явным спуском по типам AST, который **не заходит** в `*ast.LoopStatement`.

---

## Сборка из исходников

```bash
git clone https://github.com/EvilMilko/Obfuscator-1C.git
cd Obfuscator-1C
go build -o 1c-obfuscator.exe .
```

---

## Благодарности

- [LazarenkoA/Obfuscator-1C](https://github.com/LazarenkoA/Obfuscator-1C) — оригинальный проект
- [LazarenkoA/1c-language-parser](https://github.com/LazarenkoA/1c-language-parser) — AST-парсер BSL

## Лицензия

MIT — наследуется от [оригинального репозитория](https://github.com/LazarenkoA/Obfuscator-1C).
