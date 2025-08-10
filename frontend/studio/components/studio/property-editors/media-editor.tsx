import { FC } from 'react';
import { BasePropertyEditor } from './base-property-editor';
import { CanvasItem } from '@/types/quiz';
import { RJSFSchema, UiSchema } from '@rjsf/utils';

interface MediaEditorProps {
  item: CanvasItem;
  className?: string;
}

const mediaSchema: RJSFSchema = {
  type: 'object',
  title: 'Media Block Properties',
  properties: {
    title: {
      type: 'string',
      title: 'Block Title',
      description: 'Title or caption for the media block',
      minLength: 1,
      maxLength: 500
    },
    required: {
      type: 'boolean',
      title: 'Required Interaction',
      description: 'Users must interact with media to proceed',
      default: false
    },
    points: {
      type: 'number',
      title: 'Points',
      description: 'Points awarded for viewing/interacting with media (optional)',
      minimum: 0,
      maximum: 1000
    },
    explanation: {
      type: 'string',
      title: 'Description',
      description: 'Additional context or instructions for the media',
      maxLength: 1000
    },
    content: {
      type: 'object',
      title: 'Media Configuration',
      properties: {
        url: {
          type: 'string',
          title: 'Media URL',
          description: 'URL or path to the media file',
          format: 'uri'
        },
        mediaType: {
          type: 'string',
          title: 'Media Type',
          description: 'Type of media content',
          enum: ['image', 'video', 'audio'],
          enumNames: ['Image', 'Video', 'Audio'],
          default: 'image'
        },
        altText: {
          type: 'string',
          title: 'Alt Text',
          description: 'Alternative text for accessibility (required for images)',
          maxLength: 200
        },
        caption: {
          type: 'string',
          title: 'Caption',
          description: 'Caption displayed below the media',
          maxLength: 500
        },
        autoplay: {
          type: 'boolean',
          title: 'Autoplay',
          description: 'Start playing automatically (video/audio only)',
          default: false
        },
        showControls: {
          type: 'boolean',
          title: 'Show Controls',
          description: 'Display media controls (video/audio only)',
          default: true
        },
        loop: {
          type: 'boolean',
          title: 'Loop',
          description: 'Loop media playback (video/audio only)',
          default: false
        },
        muted: {
          type: 'boolean',
          title: 'Muted',
          description: 'Start with audio muted (video/audio only)',
          default: false
        },
        width: {
          type: 'number',
          title: 'Width (pixels)',
          description: 'Custom width for the media (optional)',
          minimum: 50,
          maximum: 2000
        },
        height: {
          type: 'number',
          title: 'Height (pixels)',
          description: 'Custom height for the media (optional)',
          minimum: 50,
          maximum: 2000
        },
        aspectRatio: {
          type: 'string',
          title: 'Aspect Ratio',
          description: 'Maintain specific aspect ratio',
          enum: ['', '16:9', '4:3', '1:1', '3:2', '9:16'],
          enumNames: ['Auto', '16:9 (Widescreen)', '4:3 (Standard)', '1:1 (Square)', '3:2 (Photo)', '9:16 (Portrait)'],
          default: ''
        }
      },
      required: ['url', 'mediaType']
    }
  },
  required: ['title', 'required', 'content']
};

const mediaUiSchema: UiSchema = {
  title: {
    'ui:placeholder': 'Enter media title or caption...'
  },
  explanation: {
    'ui:widget': 'textarea',
    'ui:options': {
      rows: 2
    },
    'ui:placeholder': 'Describe the media or provide instructions...'
  },
  content: {
    'ui:title': 'Media Settings',
    'ui:order': ['url', 'mediaType', 'altText', 'caption', 'aspectRatio', 'width', 'height', 'autoplay', 'showControls', 'loop', 'muted'],
    url: {
      'ui:placeholder': 'https://example.com/media.jpg or upload://file.jpg'
    },
    mediaType: {
      'ui:help': 'Select the type of media you are embedding'
    },
    altText: {
      'ui:placeholder': 'Describe what is shown in the image...',
      'ui:help': 'Important for accessibility and when images fail to load'
    },
    caption: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 2
      },
      'ui:placeholder': 'Caption displayed below the media...'
    },
    aspectRatio: {
      'ui:help': 'Choose a ratio or leave auto to use original dimensions'
    },
    width: {
      'ui:help': 'Leave empty to use responsive sizing'
    },
    height: {
      'ui:help': 'Leave empty to maintain aspect ratio'
    },
    autoplay: {
      'ui:help': 'Note: Most browsers block autoplay with audio'
    },
    showControls: {
      'ui:help': 'Allow users to play/pause and control volume'
    },
    loop: {
      'ui:help': 'Restart media from beginning when it ends'
    },
    muted: {
      'ui:help': 'Start without sound (can be unmuted by user)'
    }
  }
};

export const MediaEditor: FC<MediaEditorProps> = ({ item, className }) => {
  return (
    <BasePropertyEditor
      item={item}
      schema={mediaSchema}
      uiSchema={mediaUiSchema}
      className={className}
    />
  );
};