import { FC, useMemo } from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import { RJSFSchema, UiSchema } from '@rjsf/utils';
import { CanvasItem, ItemType } from '@/types/quiz';
import { useCanvasStore } from '@/store/canvas-store';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { Save, RotateCcw, AlertTriangle } from 'lucide-react';

interface BasePropertyEditorProps {
  item: CanvasItem;
  schema: RJSFSchema;
  uiSchema?: UiSchema;
  className?: string;
}

// Custom field template for better styling
const FieldTemplate = (props: any) => {
  const { id, label, help, required, description, errors, children } = props;
  
  return (
    <div className="mb-4">
      {label && (
        <label htmlFor={id} className="block text-sm font-medium text-gray-700 mb-1">
          {label}
          {required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      {description && (
        <p className="text-xs text-gray-500 mb-2">{description}</p>
      )}
      {children}
      {errors && errors.length > 0 && (
        <div className="mt-1">
          {errors.map((error: string, i: number) => (
            <div key={i} className="flex items-center gap-1 text-xs text-red-600">
              <AlertTriangle className="w-3 h-3" />
              {error}
            </div>
          ))}
        </div>
      )}
      {help && (
        <div className="mt-1 text-xs text-gray-500">{help}</div>
      )}
    </div>
  );
};

// Custom object field template
const ObjectFieldTemplate = (props: any) => {
  const { title, description, properties, required, uiSchema } = props;
  
  return (
    <div className="space-y-4">
      {title && title !== 'root' && (
        <h3 className="text-sm font-semibold text-gray-800 border-b border-gray-200 pb-2">
          {title}
        </h3>
      )}
      {description && (
        <p className="text-xs text-gray-600">{description}</p>
      )}
      <div className="space-y-3">
        {properties.map((element: any) => element.content)}
      </div>
    </div>
  );
};

// Custom array field template
const ArrayFieldTemplate = (props: any) => {
  const { title, items, canAdd, onAddClick } = props;
  
  return (
    <div className="space-y-3">
      {title && (
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-medium text-gray-700">{title}</h4>
          {canAdd && (
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={onAddClick}
              className="text-xs"
            >
              Add Item
            </Button>
          )}
        </div>
      )}
      <div className="space-y-2">
        {items.map((element: any, index: number) => (
          <div key={index} className="relative border border-gray-200 rounded-lg p-3">
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                {element.children}
              </div>
              {element.hasMoveUp && (
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={element.onReorderClick(index, index - 1)}
                  className="p-1 h-auto"
                >
                  ↑
                </Button>
              )}
              {element.hasMoveDown && (
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={element.onReorderClick(index, index + 1)}
                  className="p-1 h-auto"
                >
                  ↓
                </Button>
              )}
              {element.hasRemove && (
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  onClick={element.onDropIndexClick(index)}
                  className="p-1 h-auto text-red-600 hover:text-red-700"
                >
                  ×
                </Button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Custom widgets for better UX
const widgets = {
  CheckboxWidget: (props: any) => (
    <label className="flex items-center gap-2 cursor-pointer">
      <input
        type="checkbox"
        checked={props.value || false}
        onChange={(e) => props.onChange(e.target.checked)}
        className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        disabled={props.disabled}
      />
      <span className="text-sm text-gray-700">{props.label}</span>
    </label>
  ),
  
  TextWidget: (props: any) => (
    <input
      type="text"
      value={props.value || ''}
      onChange={(e) => props.onChange(e.target.value)}
      placeholder={props.placeholder}
      className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      disabled={props.disabled}
    />
  ),
  
  TextareaWidget: (props: any) => (
    <textarea
      value={props.value || ''}
      onChange={(e) => props.onChange(e.target.value)}
      placeholder={props.placeholder}
      rows={props.rows || 3}
      className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none"
      disabled={props.disabled}
    />
  ),
  
  SelectWidget: (props: any) => (
    <select
      value={props.value || ''}
      onChange={(e) => props.onChange(e.target.value)}
      className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      disabled={props.disabled}
    >
      {props.options.enumOptions?.map((option: any, i: number) => (
        <option key={i} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ),
  
  NumberWidget: (props: any) => (
    <input
      type="number"
      value={props.value || ''}
      onChange={(e) => props.onChange(e.target.value ? Number(e.target.value) : undefined)}
      placeholder={props.placeholder}
      min={props.schema.minimum}
      max={props.schema.maximum}
      className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      disabled={props.disabled}
    />
  )
};

export const BasePropertyEditor: FC<BasePropertyEditorProps> = ({
  item,
  schema,
  uiSchema,
  className
}) => {
  const { updateItem, saveState } = useCanvasStore();

  // Merge basic item properties with type-specific content
  const formData = useMemo(() => ({
    title: item.title,
    required: item.required,
    points: item.points,
    explanation: item.explanation,
    content: item.content || {}
  }), [item]);

  const handleSubmit = ({ formData }: any) => {
    const { content, ...itemProperties } = formData;
    
    updateItem(item.id, {
      ...itemProperties,
      content: content
    });
    
    saveState();
  };

  const handleChange = ({ formData }: any) => {
    // Real-time updates without saving to history
    const { content, ...itemProperties } = formData;
    
    updateItem(item.id, {
      ...itemProperties,
      content: content
    });
  };

  const handleReset = () => {
    // Reset to original values would require implementing form reset
    // For now, we'll just trigger a re-render by updating with current values
    updateItem(item.id, {
      title: item.title,
      required: item.required,
      points: item.points,
      explanation: item.explanation,
      content: item.content
    });
  };

  return (
    <div className={cn('space-y-4', className)}>
      <Form
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onChange={handleChange}
        onSubmit={handleSubmit}
        validator={validator}
        widgets={widgets}
        templates={{
          FieldTemplate,
          ObjectFieldTemplate,
          ArrayFieldTemplate
        }}
        showErrorList={false}
        liveValidate={true}
        noHtml5Validate={true}
      >
        {/* Custom submit button */}
        <div className="flex gap-2 pt-4 border-t border-gray-200">
          <Button
            type="submit"
            size="sm"
            className="flex-1 gap-2"
          >
            <Save className="w-4 h-4" />
            Save Changes
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleReset}
          >
            <RotateCcw className="w-4 h-4" />
          </Button>
        </div>
      </Form>
    </div>
  );
};