# NEO Bank — Frontend: Styles

The token layer: Tailwind CSS utilities configured from `07-ui-ux-design.md`, and the DaisyUI component classes built on top of them.

Back to the [frontend index](../README.md) · [shared frontend rules](../README.md#2-shared-rules) · [tech stack decision record](../../../08-tech-stack.md).

## 1. Documents

| Document                             | Covers                                         |
| ------------------------------------ | ---------------------------------------------- |
| [`tailwindcss.md`](./tailwindcss.md) | Design tokens, utility conventions, responsive |
| [`daisyui.md`](./daisyui.md)         | Theme mapping, component classes, shared UI    |

## 2. Styles Rules

1. **`07-ui-ux-design.md` is the source of truth.** Tailwind consumes the tokens; nothing invents a value (see [`tailwindcss.md`](./tailwindcss.md)).
2. **Both themes are first-class.** Every colour resolves through the theme (`neo-light`, `neo-dark`); never hardcode a hex that only works in one theme.
3. **Component classes over utility chains.** Use DaisyUI `btn`, `card`, `badge`, `alert` rather than re-deriving them (see [`daisyui.md`](./daisyui.md)).
4. **No dynamic class names.** Class strings are static so the purge step can see them.
5. **No inline styles.** Style through the token layer only.

## 3. Token Sources

| Token category | Source document                        | Consumed by                          |
|----------------|----------------------------------------|--------------------------------------|
| Colours        | `07-ui-ux-design.md` (colour palette)  | Tailwind `theme.colors`, DaisyUI theme |
| Spacing        | `07-ui-ux-design.md` (spacing scale)   | Tailwind `theme.spacing`             |
| Border radius  | `07-ui-ux-design.md` (radius scale)    | Tailwind `theme.borderRadius`        |
| Typography     | `07-ui-ux-design.md` (type scale)      | Tailwind `theme.fontSize`, `fontFamily` |
| Breakpoints    | `07-ui-ux-design.md` (responsive)      | Tailwind `theme.screens`             |

## 4. Semantic Colour Mapping

| Semantic role | Light theme | Dark theme | Usage                        |
|---------------|-------------|------------|------------------------------|
| Primary       | `#2563EB`   | `#3B82F6`  | Buttons, links, active states |
| Success       | `#10B981`   | `#34D399`  | Credits, confirmations        |
| Error         | `#EF4444`   | `#F87171`  | Debits, failures, validation  |
| Warning       | `#F59E0B`   | `#FBBF24`  | Pending transfers             |
| Info          | `#3B82F6`   | `#60A5FA`  | Informational messages        |
| Text primary  | `#111827`   | `#F9FAFB`  | Body text, headings           |
| Text secondary| `#6B7280`   | `#9CA3AF`  | Helper text, captions         |
| Border        | `#E5E7EB`   | `#374151`  | Input borders, dividers       |

## 5. Definition of Done (Styles)

1. No raw hex or arbitrary values outside the theme definition.
2. Every surface, text, and state colour is legible in both light and dark themes.
3. Interactive components keep visible focus indicators and 44px touch targets.
4. Component classes (DaisyUI) preferred over long utility chains.
5. No dynamic class name construction.

## 6. Cross-references

- UI/UX design: [`../../../07-ui-ux-design.md`](../../../07-ui-ux-design.md)
- Tailwind practices: [`./tailwindcss.md`](./tailwindcss.md)
- DaisyUI practices: [`./daisyui.md`](./daisyui.md)
- Component rules: [`../core/react.md`](../core/react.md)
- Planned test: `test_frontend_styles_index` (verifies styles documents exist and cross-reference correctly)

---

↑ Part of the [NEO Bank tech stack decision record](../../../08-tech-stack.md).