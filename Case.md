# КЕЙС 1. «ПРОЕКТ ПОД КОНТРОЛЕМ»

## Разработка приложения для управления проектами и построения диаграммы Ганта

## Контекст

При работе над проектом команде необходимо понимать, какие задачи предстоит выполнить, кто за них отвечает, сколько времени они займут и как одна задача влияет на другую.

Чем больше задач в проекте, тем сложнее контролировать сроки и быстро оценивать текущее состояние проекта.

Существующие системы управления проектами часто перегружены функциями и требуют времени на освоение.

## Бизнес-задача

Создайте MVP веб-приложения, которое помогает пользователю планировать проект, управлять задачами и контролировать сроки с помощью диаграммы Ганта.

Приложение должно позволять быстро создать план проекта и ответить на главный вопросы:

Какие задачи есть?

Кто отвечает за задачи?

Какие сроки начала и конца у задач?

Как задачи зависят друг от друга?

Если мы сместим выполнение задачи (например из-за отпуска сотрудника), то как это повлияет на проект.

## Что необходимо реализовать

## 1. Создание проекта

Пользователь должен иметь возможность:

создать проект

указать название и сроки проекта

добавить задачи

указать срок выполнения задачи

назначить ответственного

установить статус задачи

## 2. Диаграмма Ганта

Для задач необходимо визуально отобразить:

дату начала

дату окончания

продолжительность

статус

взаимосвязи между задачами

Пользователь должен легко понимать последовательность работ и текущее состояние проекта.

## 3. Зависимости между задачами

Пользователь должен иметь возможность связать задачи.

Например:

Анализ требований -> Дизайн -> Разработка -> Тестирование

При изменении сроков пользователь должен понимать, какие последующие задачи могут быть затронуты.



## Что необходимо реализовать

## 4. Управление проектом

Необходимо реализовать возможность:

изменить сроки

изменить статус

изменить ответственного

добавить или удалить задачу

отредактировать задачу

## 5. Контроль состояния

Приложение должно помогать быстро определить:

какие задачи выполнены

какие выполняются

какие просрочены

какие предстоит выполнить

есть ли риск нарушения общего срока проекта

## Дополнительные возможности

Дополнительный функционал остается на выбор команды.

Например:

фильтрация задач

поиск

процент выполнения проекта

выделение критического пути

уведомления о просроченных задачах

комментарии к задачам

несколько представлений проекта

экспорт проекта

совместная работа нескольких пользователей

### Дополнительные функции должны усиливать основной сценарий, а не заменять его.

## Ограничения

Не требуется создавать полноценную систему управления проектами.

Ваша задача — разработать работающий MVP<sub>,</sub> демонстрирующий удобный способ планирования и контроля проекта.

## Критерии оценки полуфинала

## Максимальное количество — 30 баллов.

По каждому критерию команда получает от 0 до 5 баллов.

<table style="min-width: 300px;"><colgroup><col style="min-width: 100px;"><col style="min-width: 100px;"><col style="min-width: 100px;"></colgroup><tbody><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>№</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Критерий</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Что оцениваем</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>1</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Работоспособность основного сценария</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Можно ли полностью пройти основной пользовательский путь от начала до результата без критических ошибок</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>2</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Корректность работы</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько корректно система обрабатывает действия пользователя, данные и формирует ожидаемый результат</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>3</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Архитектурное решение</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько выбранная архитектура соответствует поставленной задаче, обеспечивает устойчивость решения и возможность его дальнейшего развития</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>4</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Качество реализации</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько последовательно и качественно реализована функциональность: структура решения, организация кода, обработка ошибок, отсутствие критических технических проблем</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>5</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Полнота реализации кейса</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько команда реализовала обязательные требования кейса и ключевой пользовательский сценарий</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>6</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Техническая проработка дополнительных возможностей</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько дополнительные функции реализованы качественно и действительно усиливают основное решение</p></td></tr></tbody></table>

## Критерии оценки финала

## Максимальное количество — 30 баллов.

По каждому критерию команда получает от 0 до 5 баллов.

<table style="min-width: 300px;"><colgroup><col style="min-width: 100px;"><col style="min-width: 100px;"><col style="min-width: 100px;"></colgroup><tbody><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>№</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Критерий</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Что оцениваем</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>1</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Качество продукта</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько решение соответствует задаче кейса, насколько полно реализован основной сценарий и насколько результат соответствует ожиданиям целевого пользователя</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>2</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Пользовательский интерфейс и удобство</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько интерфейс понятен, логичен и удобен; легколи пользователю разобраться в продукте и выполнить нужные действия</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>3</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Качество демонстрации</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько наглядно команда показывает работу продукта: последовательно ли демонстрирует основной сценарий, понятно ли объясняет возможности и результат</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>4</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Продуктовые решения</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько команда понимает потребности пользователя и может объяснить, почему выбрала именно такой сценарий, функциональность и подход к решению задачи</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>5</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Командная работа</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько эффективно участники распределили роли и задачи, понимают вклад друг друга и могут объяснить, как организовали совместную работу над решением</p></td></tr><tr><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>6</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Ответы на вопросы</p></td><td colspan="1" rowspan="1" data-border-top="true" data-border-right="true" data-border-bottom="true" data-border-left="true" text-align="left" style="text-align: left;"><p>Насколько команда уверенно и по существу отвечает на вопросы жюри, понимает ограничения и сильные стороны своего решения и может обосновать принятые решения</p></td></tr></tbody></table>