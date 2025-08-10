import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface TextEntryEditorProps {
  item: CanvasItem;
  className?: string;
}

const textEntrySchema: RJSFSchema = {
  type: 'object',
  title: 'Text Entry Question Properties',
  properties: {
    title: {
      type: 'string',
      title: 'Question Text',
      description: 'The main question text that users will see',
      minLength: 1,
      maxLength: 500
    },
    required: {
      type: 'boolean',
      title: 'Required Question',
      description: 'Users must answer this question to proceed',
      default: false
    },
    points: {
      type: 'number',
      title: 'Points',
      description: 'Points awarded for correct answer (optional)',
      minimum: 0,
      maximum: 1000
    },
    explanation: {
      type: 'string',
      title: 'Explanation',
      description: 'Feedback shown after answering (optional)',
      maxLength: 1000
    },
    content: {
      type: 'object',
      title: 'Input Configuration',
      properties: {
        maxLength: {
          type: 'number',
          title: 'Maximum Length',
          description: 'Maximum number of characters allowed (optional)',
          minimum: 1,
          maximum: 10000
        },
        placeholder: {
          type: 'string',
          title: 'Placeholder Text',
          description: 'Hint text shown when input is empty',
          maxLength: 100
        },
        multiline: {
          type: 'boolean',
          title: 'Multi-line Input',
          description: 'Allow multiple lines of text (textarea)',
          default: false
        },
        correctAnswer: {
          type: 'string',
          title: 'Sample Correct Answer',
          description: 'Example of a correct answer for auto-grading (optional)',
          maxLength: 10000
        },
        caseSensitive: {
          type: 'boolean',
          title: 'Case Sensitive',
          description: 'Whether answer matching should be case sensitive',
          default: false
        },
        exactMatch: {
          type: 'boolean',
          title: 'Exact Match Required',
          description: 'Require exact match or allow partial matching',
          default: false
        },
        acceptableAnswers: {
          type: 'array',
          title: 'Additional Acceptable Answers',
          description: 'Alternative correct answers (for auto-grading)',
          items: {
            type: 'string',
            title: 'Answer',
            maxLength: 10000
          },
          maxItems: 10
        }
      }
    }
  },
  required: ['title', 'required', 'content']
};

const textEntryUiSchema: UiSchema = {
  title: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 2
    }
  },
  explanation: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 3
    }
  },
  content: {
    'ui:title': 'Input Settings',
    'ui:order': ['placeholder', 'multiline', 'maxLength', 'correctAnswer', 'caseSensitive', 'exactMatch', 'acceptableAnswers'],
    placeholder: {
      'ui:placeholder': 'e.g., Enter your answer here...'
    },
    multiline: {
      'ui:help': 'Use textarea instead of single-line input'
    },
    maxLength: {
      'ui:help': 'Leave empty for no length limit'
    },
    correctAnswer: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 2
      },
      'ui:help': 'Used for automatic grading and feedback'
    },
    caseSensitive: {
      'ui:help': 'When disabled, "Answer" and "answer" are treated the same'
    },
    exactMatch: {
      'ui:help': 'When disabled, allows partial matching and trimming whitespace'
    },
    acceptableAnswers: {
      'ui:help': 'Alternative ways to express the correct answer',
      items: {
        'ui:placeholder': 'Alternative answer...'
      }
    }
  }
};

export const TextEntryEditor: FC<TextEntryEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={textEntrySchema}
      uiSchema={textEntryUiSchema}
      className={className}
    />
  );
};