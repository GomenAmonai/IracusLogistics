// Телефон/почта появятся с реальными после регистрации юрлица — фейковые не публикуем.
const CONTACTS = [
  { label: 'Менеджер', value: '@hikill8', href: 'https://t.me/hikill8' },
  { label: 'Бот', value: '@IcarisLogBot', href: 'https://t.me/IcarisLogBot' },
]

export function Footer() {
  return (
    <footer id="contacts" className="border-t border-line bg-base">
      <div className="mx-auto w-full max-w-6xl px-5 py-16 sm:px-8">
        <div className="grid gap-10 md:grid-cols-[1.35fr_1fr_auto]">
          <div>
            {/* Бренд-марка как в шапке: кобальтовый скруглённый квадрат-литера + Icaris */}
            <div className="flex items-center gap-2.5">
              <span
                aria-hidden="true"
                className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent font-display text-sm font-bold text-surface"
              >
                I
              </span>
              <span className="font-display text-lg font-extrabold tracking-[-0.01em] text-ink">
                Icaris
              </span>
            </div>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-soft">
              Организация доставки грузов из Китая в Россию: расчёт маршрута, сопровождение
              и связь с менеджером в одном сервисе.
            </p>
          </div>

          <div className="border-y border-line-soft py-4 text-sm leading-relaxed text-ink-soft">
            <p className="font-medium text-ink">Юридическая информация</p>
            <p className="mt-2">
              Реквизиты стороны договора и согласованные условия предоставляются клиенту до
              начала работ и оплаты.
            </p>
          </div>

          <div className="flex flex-col gap-5">
            <ul className="space-y-3 text-sm">
              {CONTACTS.map((contact) => (
                <li key={contact.label} className="flex flex-col">
                  <span className="text-[0.7rem] uppercase tracking-[0.06em] text-ink-soft">
                    {contact.label}
                  </span>
                  <a
                    href={contact.href}
                    className="inline-flex min-h-11 items-center font-mono text-ink underline-offset-4 transition-colors duration-200 hover:text-accent hover:underline"
                  >
                    {contact.value}
                  </a>
                </li>
              ))}
            </ul>
            <a
              href="#lead"
              className="inline-flex items-center justify-center rounded-full bg-accent px-6 py-3 text-sm font-semibold text-surface shadow-card transition-colors duration-200 hover:bg-accent-deep"
            >
              Запросить расчёт
            </a>
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-2 border-t border-line pt-6 text-xs text-ink-soft sm:flex-row sm:items-center sm:justify-between">
          <span>© {new Date().getFullYear()} Icaris</span>
          <span>Грузы под контролем</span>
        </div>
      </div>
    </footer>
  )
}
