import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface TitleEditorProps {
  item: CanvasItem;
  className?: string;
}

const titleSchema: RJSFSchema = {
  type: 'object',
  title: 'Title Block Properties',
  properties: {
    title: {
      type: 'string',
      title: 'Title Text',
      description: 'The main heading or title text',
      minLength: 1,
      maxLength: 500
    },
    required: {
      type: 'boolean',
      title: 'Acknowledgment Required',
      description: 'Users must acknowledge reading this title to proceed',
      default: false
    },
    explanation: {
      type: 'string',
      title: 'Subtitle or Description',
      description: 'Additional text shown below the title (optional)',
      maxLength: 1000
    },
    content: {
      type: 'object',
      title: 'Formatting Options',
      properties: {
        level: {
          type: 'number',
          title: 'Heading Level',
          description: 'HTML heading level (affects text size and hierarchy)',
          enum: [1, 2, 3, 4, 5, 6],
          enumNames: ['H1 (Largest)', 'H2', 'H3', 'H4', 'H5', 'H6 (Smallest)'],
          default: 2
        },
        alignment: {
          type: 'string',
          title: 'Text Alignment',
          description: 'How to align the title text',
          enum: ['left', 'center', 'right'],
          enumNames: ['Left', 'Center', 'Right'],
          default: 'left'
        },
        color: {
          type: 'string',
          title: 'Text Color',
          description: 'Color of the title text',
          enum: ['default', 'primary', 'secondary', 'success', 'warning', 'error'],
          enumNames: ['Default', 'Primary Blue', 'Secondary Gray', 'Success Green', 'Warning Orange', 'Error Red'],
          default: 'default'
        },
        emphasis: {
          type: 'string',
          title: 'Text Style',
          description: 'Visual emphasis for the title',
          enum: ['normal', 'bold', 'italic', 'underline'],
          enumNames: ['Normal', 'Bold', 'Italic', 'Underlined'],
          default: 'bold'
        },
        showDivider: {
          type: 'boolean',
          title: 'Show Divider Line',
          description: 'Display a horizontal line below the title',
          default: false
        },
        dividerColor: {
          type: 'string',
          title: 'Divider Color',
          description: 'Color of the divider line',
          enum: ['gray', 'primary', 'secondary'],
          enumNames: ['Gray', 'Primary', 'Secondary'],
          default: 'gray'
        }
      }
    }
  },
  required: ['title', 'required', 'content']
};

const titleUiSchema: UiSchema = {
  title: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 2
    },
    'ui:placeholder': 'Enter your title or heading...'
  },
  explanation: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 3
    },
    'ui:placeholder': 'Optional subtitle or description...'
  },
  content: {
    'ui:title': 'Styling Options',
    'ui:order': ['level', 'alignment', 'color', 'emphasis', 'showDivider', 'dividerColor'],
    level: {
      'ui:help': 'H1 is largest, H6 is smallest. Use H1 sparingly for main titles.'
    },
    alignment: {
      'ui:help': 'How the title appears on the page'
    },
    color: {
      'ui:help': 'Choose a color that fits your quiz theme'
    },
    emphasis: {
      'ui:help': 'Make the title stand out with formatting'
    },
    showDivider: {
      'ui:help': 'Adds a horizontal line to separate sections'
    },
    dividerColor: {
      'ui:help': 'Only applies when divider is enabled'
    }
  }
};

export const TitleEditor: FC<TitleEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={titleSchema}
      uiSchema={titleUiSchema}
      className={className}
    />
  );
};