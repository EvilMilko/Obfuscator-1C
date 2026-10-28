package obfuscator

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSergeyReportedCases(t *testing.T) {
	// Кейс 1: Группировки СКД и присваивание булево
	codeGrouping := `&НаСервере
Процедура СоздатьГруппировку()
	НовГруппировка = Список.КомпоновщикНастроек.Настройки.Структура.Добавить(Тип("ГруппировкаКомпоновкиДанных"));
	НовГруппировка.Имя = "ВладелецПартии";
	НовГруппировка.Использование = Истина;
КонецПроцедуры`

	cfg := Config{
		RepExpByTernary:  true,
		RepLoopByGoto:    true,
		RepExpByEval:     true,
		HideString:       true,
		ChangeConditions: true,
		AppendGarbage:    true,
		CallStackHell:    true,
		LineBreaks:       true,
	}

	obf := NewObfuscatory(context.Background(), cfg)
	resGrouping, err := obf.Obfuscate(codeGrouping)
	assert.NoError(t, err)

	// Проверяем отсутствие битых символов \ufffd
	assert.NotContains(t, resGrouping, "\ufffd")
	// Проверяем, что директива компиляции у вспомогательной функции корректна
	if strings.Contains(resGrouping, "_ТранзакцияАктивна") {
		assert.Contains(t, resGrouping, "&НаКлиентеНаСервереБезКонтекста")
	}

	// Кейс 2: Строка запроса с кавычками и многострочным |
	codeQuery := `&НаСервере
Процедура ТестЗапрос()
	Текст = "КОГДА СправочникПартииНоменклатуры.ВладелецПартии = ЗНАЧЕНИЕ(Справочник.Контрагенты.ПустаяСсылка)
	| ТОГДА ""<Без владельца>""";
КонецПроцедуры`

	obfQuery := NewObfuscatory(context.Background(), cfg)
	resQuery, err := obfQuery.Obfuscate(codeQuery)
	assert.NoError(t, err)
	assert.NotContains(t, resQuery, "\ufffd")

	// Кейс 3: NoEvalFuncs исключает _ТранзакцияАктивна
	cfgNoEval := Config{
		RepExpByEval: true,
		NoEvalFuncs:  []string{"СоздатьГруппировку"},
	}
	obfNoEval := NewObfuscatory(context.Background(), cfgNoEval)
	resNoEval, err := obfNoEval.Obfuscate(codeGrouping)
	assert.NoError(t, err)
	assert.NotContains(t, resNoEval, "_ТранзакцияАктивна")
	assert.NotContains(t, resNoEval, "Выполнить")
	assert.NotContains(t, resNoEval, "Вычислить")
}
