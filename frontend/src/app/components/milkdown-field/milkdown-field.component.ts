import {
  Component, Input, Output, EventEmitter,
  ElementRef, ViewChild, AfterViewInit, OnDestroy,
} from '@angular/core';
import Vditor from 'vditor';

@Component({
  selector: 'app-milkdown-field',
  standalone: true,
  template: `<div #editorContainer class="vditor-container"></div>`,
  styleUrl: './milkdown-field.component.scss',
})
export class VditorFieldComponent implements AfterViewInit, OnDestroy {
  @Input() value = '';
  @Output() valueChange = new EventEmitter<string>();

  @ViewChild('editorContainer', { static: true }) container!: ElementRef<HTMLDivElement>;

  private vditor?: Vditor;

  ngAfterViewInit(): void {
    setTimeout(() => {
      this.vditor = new Vditor(this.container.nativeElement, {
        height: 280,
        mode: 'wysiwyg',
        placeholder: '请输入内容…',
        value: this.value || '',
        cache: {
          enable: false,
          id: `${Date.now()}-${Math.random().toString(36).slice(2, 11)}`,
        },
      toolbar: [
        'headings', 'bold', 'italic', 'strike', '|',
        'list', 'ordered-list', 'check', '|',
        'quote', 'line', 'code', 'inline-code', '|',
        'table', 'link', 'upload', '|',
        'undo', 'redo', '|',
        'preview', 'fullscreen',
        { name: 'export', tipPosition: 's' },
      ],
      upload: {
        url: '/api/upload',
        max: 10 * 1024 * 1024,
        accept: 'image/*',
        fieldName: 'file',
      },
      counter: { enable: true, max: 200000 },
        input: (val: string) => {
          this.valueChange.emit(val);
        },
      });
    });
  }

  ngOnDestroy(): void {
    this.vditor?.destroy();
  }
}
