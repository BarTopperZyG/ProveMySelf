import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface ChoiceEditorProps {
  item: CanvasItem;
  className?: string;
}

const choiceSchema: RJSFSchema = {
  type: 'object',
  title: 'Choice Question Properties',
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
      title: 'Answer Choices',
      properties: {
        choices: {
          type: 'array',
          title: 'Choices',
          description: 'List of answer choices. At least one must be marked correct.',
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
                description: 'Mark this choice as the correct answer',
                default: false
              }
            },
            required: ['id', 'text']
          },
          default: [
            { id: 'choice_1', text: 'Option A', correct: true },
            { id: 'choice_2', text: 'Option B', correct: false },
            { id: 'choice_3', text: 'Option C', correct: false },
            { id: 'choice_4', text: 'Option D', correct: false }
          ]
        }
      },
      required: ['choices']
    }
  },
  required: ['title', 'required', 'content']
};

const choiceUiSchema: UiSchema = {
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
          'ui:help': 'Only one choice should be marked correct for single-choice questions'
        }
      }
    }
  }
};

export const ChoiceEditor: FC<ChoiceEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={choiceSchema}
      uiSchema={choiceUiSchema}
      className={className}
    />
  );
};