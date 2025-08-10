import { FC } from 'react';
import { useDraggable } from '@dnd-kit/core';
import { ItemType } from '@/types/quiz';
import { cn } from '@/lib/utils';
import { 
  Type, 
  Image, 
  CheckCircle, 
  CheckSquare, 
  FileText, 
  List, 
  MousePointer 
} from 'lucide-react';

interface BlockPaletteProps {
  className?: string;
}

interface BlockItemProps {
  type: ItemType;
  icon: React.ReactNode;
  title: string;
  description: string;
}

const BLOCK_ITEMS: BlockItemProps[] = [
  {
    type: ItemType.TITLE,
    icon: <Type className="w-5 h-5" />,
    title: 'Title Block',
    description: 'Add headings and instructional text'
  },
  {
    type: ItemType.MEDIA,
    icon: <Image className="w-5 h-5" />,
    title: 'Media Block',
    description: 'Embed images, videos, or audio'
  },
  {
    type: ItemType.CHOICE,
    icon: <CheckCircle className="w-5 h-5" />,
    title: 'Single Choice',
    description: 'Multiple choice, single answer'
  },
  {
    type: ItemType.MULTI_CHOICE,
    icon: <CheckSquare className="w-5 h-5" />,
    title: 'Multiple Choice',
    description: 'Multiple choice, multiple answers'
  },
  {
    type: ItemType.TEXT_ENTRY,
    icon: <FileText className="w-5 h-5" />,
    title: 'Text Entry',
    description: 'Free text input questions'
  },
  {
    type: ItemType.ORDERING,
    icon: <List className="w-5 h-5" />,
    title: 'Ordering',
    description: 'Drag and drop to order items'
  },
  {
    type: ItemType.HOTSPOT,
    icon: <MousePointer className="w-5 h-5" />,
    title: 'Hotspot',
    description: 'Click areas on images'
  }
];

const DraggableBlockItem: FC<BlockItemProps> = ({ type, icon, title, description }) => {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: `palette-${type}`,
    data: {
      type,
      isNew: true
    }
  });

  const style = transform ? {
    transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
  } : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        'flex flex-col items-center p-4 border-2 border-dashed border-gray-300 rounded-lg cursor-grab',
        'hover:border-blue-400 hover:bg-blue-50 transition-colors duration-200',
        'group select-none',
        isDragging && 'opacity-50 cursor-grabbing'
      )}
    >
      <div className="flex items-center justify-center w-10 h-10 mb-2 bg-gray-100 rounded-full group-hover:bg-blue-100 transition-colors">
        {icon}
      </div>
      <h3 className="text-sm font-medium text-gray-900 mb-1 text-center">
        {title}
      </h3>
      <p className="text-xs text-gray-500 text-center leading-tight">
        {description}
      </p>
    </div>
  );
};

export const BlockPalette: FC<BlockPaletteProps> = ({ className }) => {
  return (
    <div className={cn('w-64 bg-white border-r border-gray-200', className)}>
      <div className="p-4 border-b border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900">Question Blocks</h2>
        <p className="text-sm text-gray-500 mt-1">
          Drag blocks onto the canvas to build your quiz
        </p>
      </div>
      
      <div className="p-4 space-y-4 overflow-y-auto h-full">
        {BLOCK_ITEMS.map((item) => (
          <DraggableBlockItem
            key={item.type}
            type={item.type}
            icon={item.icon}
            title={item.title}
            description={item.description}
          />
        ))}
      </div>
      
      <div className="p-4 border-t border-gray-200 bg-gray-50">
        <div className="text-xs text-gray-500 space-y-1">
          <p className="font-medium">Tips:</p>
          <p>• Drag blocks to the canvas to create questions</p>
          <p>• Click items on canvas to edit properties</p>
          <p>• Use Ctrl+Z/Ctrl+Y to undo/redo changes</p>
        </div>
      </div>
    </div>
  );
};