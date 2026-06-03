import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'
import { Decoration, DecorationSet } from '@tiptap/pm/view'

export const spellCheckPluginKey = new PluginKey('spellCheck')

/**
 * TipTap Extension that manages spell-check decorations via
 * ProseMirror's Plugin / DecorationSet API.
 */
export const SpellCheckExtension = Extension.create({
  name: 'spellCheck',

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: spellCheckPluginKey,

        state: {
          init() {
            return DecorationSet.empty
          },

          apply(tr, decorationSet) {
            // Check if this transaction carries new spell-check decorations
            const meta = tr.getMeta(spellCheckPluginKey)
            if (meta && meta.decorations) {
              return meta.decorations
            }

            // Map existing decorations through document changes
            if (tr.docChanged) {
              return decorationSet.map(tr.mapping, tr.doc)
            }

            return decorationSet
          },
        },

        props: {
          decorations(state) {
            return spellCheckPluginKey.getState(state)
          },
        },
      }),
    ]
  },
})

/**
 * Apply spell-check decorations to the editor.
 *
 * @param {import('@tiptap/core').Editor} editor
 * @param {Array<{from: number, to: number, word: string}>} misspelledRanges
 */
export function applySpellDecorations(editor, misspelledRanges) {
  if (!editor || editor.isDestroyed) return

  const { state, view } = editor
  const decorations = misspelledRanges
    .filter(({ from, to }) => from >= 0 && to <= state.doc.content.size)
    .map(({ from, to, word }) =>
      Decoration.inline(from, to, {
        class: 'spelling-error',
        'data-word': word,
      }),
    )

  const decorationSet = DecorationSet.create(state.doc, decorations)

  const tr = state.tr.setMeta(spellCheckPluginKey, { decorations: decorationSet })
  view.dispatch(tr)
}
