# schemadoc

`schemadoc` собирает описание атрибута Terraform Plugin Framework из правил,
которые этот атрибут проверяют. Лимиты, допустимые значения, зависимости,
default и пересоздание объявляются один раз — в валидаторе или модификаторе.
Текст и порядок предложений соответствуют требованиям технических писателей
к описаниям атрибутов.

## Как описать ресурс

```go
var volumeDocs = schemadoc.Resource{Name: "network volume", Location: "network-volumes"}

Attributes: map[string]schema.Attribute{
    "id":         volumeDocs.ID(),
    "project_id": volumeDocs.ProjectID(),
    "region":     volumeDocs.Region(),
    "name": volumeDocs.String(schema.StringAttribute{
        Optional:   true,
        Computed:   true,
        Default:    schemadoc.StaticString(""),
        Validators: []validator.String{schemadoc.LengthAtMost(255), schemadoc.Trimmed()},
    }, schemadoc.Field{
        Definition: "Name of the network volume.",
    }),
},
Blocks: map[string]schema.Block{
    "timeouts": volumeDocs.Timeouts(ctx, schemadoc.Timeouts{Create: 10 * time.Minute}),
},
```

Результат `make docs`:

```markdown
- `name` (String) Name of the network volume. The maximum length is `255` characters. The value cannot start or end with whitespace. The default value is an empty string.
```

1. Объявите `schemadoc.Resource` или `schemadoc.DataSource` с существительным объекта.
   Оно попадает в «Unique identifier of the …», «Pool where the … is located» и
   «Changing this creates a new …».
   `Location` — якорь раздела продукта на странице Product availability by location,
   на него ведёт ссылка в описании `region`.
2. Общие атрибуты берите готовыми: `ID()`, `ProjectID()`, `Region()`, `Timeouts()`.
   У ресурса и data source с одним существительным описания совпадают дословно.
3. Остальные атрибуты оборачивайте в `String`, `Int64`, `Bool`, `Map`, `Set` или `List`
   **после** того, как заданы validators, default и plan modifiers.
4. В `Field` пишите только то, что не следует из правил:
   - `Definition` — первое предложение;
   - `Behavior` — особенности API и импорта;
   - `Source` — откуда взять ID или имя: `FromDataSource`, `FromResource`, `FromOpenStackCLI`.
     Несколько источников склеивайте в порядке, который требуют технические писатели:
     ресурс, data source, OpenStack CLI; уточнение к источнику (например, фильтры
     data source) — следующим предложением там же:
     в `Behavior` источник окажется раньше остальных;
   - `Concept` — ссылка «Learn more about …»; для документации вендора задайте `Vendor`, и ссылка будет в формате «For more information about …, see the [official … documentation](…)».
5. Запустите `make schemadoc` и `make docs`.

Порядок предложений задаёт `render` по требованиям технических писателей:
определение → ограничения → зависимости → Behavior → исключение из пересоздания →
Source → default → Learn more → пересоздание.

Describer'ы пишут текст в `MarkdownDescription`. Для атрибутов, описанных вручную
без describer'а, действует обычная практика Plugin Framework: текст с разметкой
(inline code, ссылки) — в `MarkdownDescription`, без разметки — в `Description`.

## Правила

| Правило | Что пишет в описание |
| --- | --- |
| `LengthAtMost`, `LengthAtLeast`, `Between`, `AtLeast` | лимит |
| `OneOf` | допустимые значения |
| `UUID`, `Trimmed`, `NonEmptyTrimmed` | формат |
| `ConflictsWith`, `ExactlyOneOf`, `AlsoRequires` | зависимость; подходят для атрибута любого вида |
| `StaticString`, `StaticBool`, `StaticInt64` | default |
| `RequiresReplace` | «Changing this creates a new …» |
| `RequiresReplaceIf` | исключение отдельным предложением и «Changing this creates a new …» последним |

`UseStateForUnknown` ничего не меняет для пользователя и используется как есть.

## Правило своего сервиса

Обычный валидатор Framework с фразой для документации:

```go
func volumeTypeIDValidator() validator.String {
    return schemadoc.DescribeString(stringvalidator.Any(
        schemadoc.UUID(),
        schemadoc.OneOf("default"),
    ), "The format is a lowercase UUID or `default` to select the default volume type.")
}
```

Для валидаторов Set и Map есть `DescribeSet` и `DescribeMap`. Перечень значений
в своём тексте собирайте через `CodeList`: он ставит Oxford comma.

Собственный тип правила встраивает `schemadoc.Facts` — этого достаточно, чтобы
реализовать `schemadoc.Documented`.

## Проверка: make schemadoc

`Lint` обходит все схемы Framework-провайдера. Схему, где нет ни одного правила
schemadoc, он пропускает: её авторы документируют атрибуты по-своему. В остальных
схемах он сообщает:

- правило без `Documentation()` — например, стандартный `stringvalidator.OneOf`
  или `stringplanmodifier.RequiresReplace`, чей смысл не попадёт в документацию;
- факт правила, которого нет в описании: атрибут описан простым `Description`,
  вид атрибута не поддержан describer'ом или правила поменяли после вызова describer'а.

Проверка входит в `make test` как `TestSchemaDocumentation`.

## Как расширить пакет

- **Новое правило** — функция в `rules.go`: стандартный валидатор HashiCorp плюс `Facts`.
- **Новый вид атрибута** (например, Float64): describer-методы у `Resource`
  и `DataSource` в `doc.go`. Правила любого вида атрибута `rulesOf` уже читает.
  Пока describer'а нет, `make schemadoc` сообщит о правилах такого атрибута.
- **Новый общий атрибут** — метод в `attributes.go` для обоих видов, чтобы
  ресурс и data source описывали его одинаково.

Тексты продуктовых особенностей и ссылок пишет автор, их проверяют на ревью
технические писатели.
