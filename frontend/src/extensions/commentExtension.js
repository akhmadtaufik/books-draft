import { Mark, mergeAttributes } from '@tiptap/core'

export const CommentMark = Mark.create({
  name: 'comment',

  addAttributes() {
    return {
      text: {
        default: null,
      },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'span[data-comment]',
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return ['span', mergeAttributes(HTMLAttributes, { 
      class: 'editor-comment', 
      'data-comment': HTMLAttributes.text 
    }), 0]
  },
})
