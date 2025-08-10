import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface MultiChoiceEditorProps {
  item: CanvasItem;
  className?: string;
}

const multiChoiceSchema: RJSFSchema = {
  type: 'object',
  title: 'Multiple Choice Question Properties',
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
      description: 'Total points for this question (distributed among correct answers)',
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
      title: 'Answer Choices',
      properties: {
        choices: {
          type: 'array',
          title: 'Choices',
          description: 'List of answer choices. Multiple can be marked correct.',
          minItems: 2,
          maxItems: 10,
          items: {
            type: 'object',
            title: 'Choice',
            properties: {
              id: {
                type: 'string',
                title: 'Choice ID',
                description: 'Unique identifier for this choice',
                default: () => `choice_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
              },
              text: {
                type: 'string',
                title: 'Choice Text',
                description: 'The text displayed for this choice',
                minLength: 1,
                maxLength: 500
              },
              correct: {
                type: 'boolean',
                title: 'Correct Answer',
                description: 'Mark this choice as one of the correct answers',
                default: false
              }
            },
            required: ['id', 'text']
          },
          default: [
            { id: 'choice_1', text: 'Option A', correct: true },
            { id: 'choice_2', text: 'Option B', correct: true },
            { id: 'choice_3', text: 'Option C', correct: false },
            { id: 'choice_4', text: 'Option D', correct: false }
          ]
        },
        partialCredit: {
          type: 'boolean',
          title: 'Allow Partial Credit',
          description: 'Award partial points for partially correct answers',
          default: true
        },
        minSelections: {
          type: 'number',
          title: 'Minimum Selections',
          description: 'Minimum number of choices users must select (optional)',
          minimum: 1,
          maximum: 10
        },
        maxSelections: {
          type: 'number',
          title: 'Maximum Selections',
          description: 'Maximum number of choices users can select (optional)',
          minimum: 1,
          maximum: 10
        }
      },
      required: ['choices']
    }
  },
  required: ['title', 'required', 'content']
};

const multiChoiceUiSchema: UiSchema = {
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
    'ui:title': 'Answer Configuration',
    'ui:order': ['choices', 'partialCredit', 'minSelections', 'maxSelections'],
    choices: {
      'ui:options': {
        addable: true,
        removable: true,
        orderable: true
      },
      items: {
        'ui:order': ['text', 'correct', 'id'],
        id: {
          'ui:widget': 'hidden'
        },
        text: {
          'ui:placeholder': 'Enter choice text...'
        },
        correct: {
          'ui:help': 'Multiple choices can be correct for multiple-choice questions'
        }
      }
    },
    partialCredit: {
      'ui:help': 'When enabled, users get partial points for selecting some but not all correct answers'
    },
    minSelections: {
      'ui:help': 'Leave empty for no minimum requirement'
    },
    maxSelections: {
      'ui:help': 'Leave empty for no maximum limit'
    }
  }
};

export const MultiChoiceEditor: FC<MultiChoiceEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={multiChoiceSchema}
      uiSchema={multiChoiceUiSchema}
      className={className}
    />
  );
};