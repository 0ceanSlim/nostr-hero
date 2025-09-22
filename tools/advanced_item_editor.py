#!/usr/bin/env python3
"""
Advanced Item Editor GUI - Browse, search, edit, and manage items from the nostr-hero game database
"""

import json
import os
import tkinter as tk
from tkinter import ttk, scrolledtext, messagebox, simpledialog
from pathlib import Path
from PIL import Image, ImageTk

class AdvancedItemEditor:
    def __init__(self, root):
        self.root = root
        self.root.title("Nostr Hero - Advanced Item Database Editor")
        self.root.geometry("1600x1000")
        
        # Load items
        self.items = self.load_items()
        self.current_item_key = None
        self.current_item_data = None
        self.sort_column = None
        self.sort_reverse = False
        
        # Create main interface
        self.create_widgets()
        self.populate_item_list()
        
    def load_items(self):
        """Load all item files from the items directory"""
        items = {}
        script_dir = Path(__file__).parent
        items_dir = script_dir.parent / "www" / "data" / "items"
        
        if not items_dir.exists():
            messagebox.showerror("Error", f"Items directory not found at {items_dir}")
            return items
        
        for item_file in items_dir.glob("*.json"):
            try:
                with open(item_file, 'r', encoding='utf-8') as f:
                    item_data = json.load(f)
                    items[item_file.stem] = item_data
            except json.JSONDecodeError as e:
                print(f"Error loading {item_file}: {e}")
            except Exception as e:
                print(f"Error reading {item_file}: {e}")
        
        return items
    
    def save_item(self, key, data):
        """Save item data back to file"""
        script_dir = Path(__file__).parent
        items_dir = script_dir.parent / "www" / "data" / "items"
        file_path = items_dir / f"{key}.json"
        
        try:
            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2)
            return True
        except Exception as e:
            messagebox.showerror("Save Error", f"Failed to save {key}.json: {e}")
            return False
    
    def save_all_items(self):
        """Save all items to their respective files"""
        success_count = 0
        for key, data in self.items.items():
            if self.save_item(key, data):
                success_count += 1
        return success_count
    
    def cp_to_gp(self, cp_value):
        """Convert copper pieces to gold pieces"""
        try:
            cp = int(cp_value) if cp_value else 0
            return cp / 100.0
        except (ValueError, TypeError):
            return 0.0
    
    def gp_to_cp(self, gp_value):
        """Convert gold pieces to copper pieces"""
        try:
            gp = float(gp_value) if gp_value else 0.0
            return int(gp * 100)
        except (ValueError, TypeError):
            return 0
    
    def create_widgets(self):
        """Create the main UI widgets"""
        # Main container
        main_frame = ttk.Frame(self.root)
        main_frame.pack(fill=tk.BOTH, expand=True, padx=10, pady=10)
        
        # Top toolbar
        toolbar_frame = ttk.Frame(main_frame)
        toolbar_frame.pack(fill=tk.X, pady=(0, 10))
        
        # Database operations
        ttk.Button(toolbar_frame, text="Add New Field to All Items", command=self.add_new_field_dialog).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(toolbar_frame, text="Remove Field from All Items", command=self.remove_field_dialog).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(toolbar_frame, text="Bulk Edit Selected", command=self.bulk_edit_dialog).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(toolbar_frame, text="Save All Changes", command=self.save_all_changes).pack(side=tk.LEFT, padx=(0, 10))
        
        # Selection info
        self.selection_label = ttk.Label(toolbar_frame, text="No items selected")
        self.selection_label.pack(side=tk.RIGHT)
        
        # Left panel - Item list and search
        left_frame = ttk.Frame(main_frame)
        left_frame.pack(side=tk.LEFT, fill=tk.BOTH, expand=False, padx=(0, 10))
        
        # Search frame
        search_frame = ttk.Frame(left_frame)
        search_frame.pack(fill=tk.X, pady=(0, 10))
        
        ttk.Label(search_frame, text="Search:").pack(anchor=tk.W)
        self.search_var = tk.StringVar()
        self.search_var.trace('w', self.on_search)
        self.search_entry = ttk.Entry(search_frame, textvariable=self.search_var)
        self.search_entry.pack(fill=tk.X, pady=(5, 0))
        
        # Filter by type
        filter_frame = ttk.Frame(left_frame)
        filter_frame.pack(fill=tk.X, pady=(0, 10))
        
        ttk.Label(filter_frame, text="Filter by Type:").pack(anchor=tk.W)
        self.type_var = tk.StringVar()
        self.type_var.trace('w', self.on_search)
        self.type_combo = ttk.Combobox(filter_frame, textvariable=self.type_var, state="readonly")
        self.type_combo.pack(fill=tk.X, pady=(5, 0))
        
        # Filter by tag
        tag_filter_frame = ttk.Frame(left_frame)
        tag_filter_frame.pack(fill=tk.X, pady=(0, 10))
        
        ttk.Label(tag_filter_frame, text="Filter by Tag:").pack(anchor=tk.W)
        self.tag_var = tk.StringVar()
        self.tag_var.trace('w', self.on_search)
        self.tag_combo = ttk.Combobox(tag_filter_frame, textvariable=self.tag_var, state="readonly")
        self.tag_combo.pack(fill=tk.X, pady=(5, 0))
        
        # Selection controls
        select_frame = ttk.Frame(left_frame)
        select_frame.pack(fill=tk.X, pady=(0, 10))
        
        ttk.Button(select_frame, text="Select All", command=self.select_all).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(select_frame, text="Clear Selection", command=self.clear_selection).pack(side=tk.LEFT)
        
        # Item list
        list_frame = ttk.Frame(left_frame)
        list_frame.pack(fill=tk.BOTH, expand=True)
        
        ttk.Label(list_frame, text="Items:").pack(anchor=tk.W)
        
        # Create treeview for item list with sorting
        tree_frame = ttk.Frame(list_frame)
        tree_frame.pack(fill=tk.BOTH, expand=True, pady=(5, 0))
        
        # Define columns
        columns = ('Type', 'Price', 'Weight', 'Stack')
        self.item_tree = ttk.Treeview(tree_frame, columns=columns, show='tree headings', selectmode='extended')
        
        # Configure headings with sorting
        self.item_tree.heading('#0', text='Name', command=lambda: self.sort_by_column('#0'))
        self.item_tree.heading('Type', text='Type', command=lambda: self.sort_by_column('Type'))
        self.item_tree.heading('Price', text='Price (gp)', command=lambda: self.sort_by_column('Price'))
        self.item_tree.heading('Weight', text='Weight', command=lambda: self.sort_by_column('Weight'))
        self.item_tree.heading('Stack', text='Stack', command=lambda: self.sort_by_column('Stack'))
        
        # Configure column widths
        self.item_tree.column('#0', width=200)
        self.item_tree.column('Type', width=150)
        self.item_tree.column('Price', width=80)
        self.item_tree.column('Weight', width=70)
        self.item_tree.column('Stack', width=60)
        
        # Scrollbars for treeview
        tree_scroll_y = ttk.Scrollbar(tree_frame, orient=tk.VERTICAL, command=self.item_tree.yview)
        tree_scroll_x = ttk.Scrollbar(tree_frame, orient=tk.HORIZONTAL, command=self.item_tree.xview)
        self.item_tree.configure(yscrollcommand=tree_scroll_y.set, xscrollcommand=tree_scroll_x.set)
        
        self.item_tree.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        tree_scroll_y.pack(side=tk.RIGHT, fill=tk.Y)
        tree_scroll_x.pack(side=tk.BOTTOM, fill=tk.X)
        
        self.item_tree.bind('<<TreeviewSelect>>', self.on_item_select)
        
        # Configure left frame width
        left_frame.configure(width=600)
        left_frame.pack_propagate(False)
        
        # Right panel - Item editor
        right_frame = ttk.Frame(main_frame)
        right_frame.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True)
        
        # Header with save button
        header_frame = ttk.Frame(right_frame)
        header_frame.pack(fill=tk.X, pady=(0, 10))
        
        ttk.Label(header_frame, text="Item Editor:", font=('TkDefaultFont', 12, 'bold')).pack(side=tk.LEFT)
        self.save_button = ttk.Button(header_frame, text="Save Changes", command=self.save_current_item)
        self.save_button.pack(side=tk.RIGHT)
        self.save_button.configure(state='disabled')
        
        # Create notebook for organized editing
        self.notebook = ttk.Notebook(right_frame)
        self.notebook.pack(fill=tk.BOTH, expand=True)
        
        # Basic info tab (includes tags and notes)
        self.basic_frame = ttk.Frame(self.notebook)
        self.notebook.add(self.basic_frame, text="Basic Info")
        self.create_basic_fields_with_image()
        
        # Raw JSON tab
        self.json_frame = ttk.Frame(self.notebook)
        self.notebook.add(self.json_frame, text="Raw JSON")
        self.create_json_field()
        
        # Status bar
        self.status_var = tk.StringVar()
        self.status_bar = ttk.Label(self.root, textvariable=self.status_var, relief=tk.SUNKEN)
        self.status_bar.pack(side=tk.BOTTOM, fill=tk.X)
    
    def sort_by_column(self, column):
        """Sort items by specified column"""
        if self.sort_column == column:
            self.sort_reverse = not self.sort_reverse
        else:
            self.sort_column = column
            self.sort_reverse = False
        
        self.refresh_item_list()
    
    def create_basic_fields_with_image(self):
        """Create basic field editors with image display"""
        # Main container with two sections: left for image, right for fields
        main_container = ttk.Frame(self.basic_frame)
        main_container.pack(fill=tk.BOTH, expand=True, padx=5, pady=5)
        
        # Left side - Image display
        image_frame = ttk.LabelFrame(main_container, text="Item Image")
        image_frame.pack(side=tk.LEFT, fill=tk.Y, padx=(0, 10))
        
        self.image_label = ttk.Label(image_frame, text="No image\navailable", justify=tk.CENTER)
        self.image_label.pack(padx=10, pady=10)
        
        # Right side - Fields
        fields_container = ttk.Frame(main_container)
        fields_container.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True)
        
        # Create scrollable frame for fields
        canvas = tk.Canvas(fields_container)
        scrollbar = ttk.Scrollbar(fields_container, orient="vertical", command=canvas.yview)
        self.scrollable_frame = ttk.Frame(canvas)
        
        self.scrollable_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        
        canvas.create_window((0, 0), window=self.scrollable_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        # Basic fields (will be populated dynamically)
        self.basic_fields = {}
        self.field_widgets = {}
        
        self.rebuild_basic_fields()
        
        canvas.pack(side="left", fill="both", expand=True)
        scrollbar.pack(side="right", fill="y")
    
    def rebuild_basic_fields(self):
        """Rebuild basic fields based on current item schema"""
        # Clear existing widgets
        for widget in self.scrollable_frame.winfo_children():
            widget.destroy()
        
        self.field_widgets = {}
        
        # Get all possible fields from all items
        all_fields = set()
        for item in self.items.values():
            all_fields.update(item.keys())
        
        # Standard field order (including tags and notes)
        standard_fields = ['name', 'description', 'price', 'type', 'weight', 'stack', 'ac', 'damage', 'heal']
        other_fields = sorted(all_fields - set(standard_fields) - {'tags', 'notes'})
        
        row = 0
        # Basic fields
        for field in standard_fields + other_fields:
            ttk.Label(self.scrollable_frame, text=f"{field.title()}:").grid(row=row, column=0, sticky='nw', padx=5, pady=2)
            
            if field == 'description':
                widget = tk.Text(self.scrollable_frame, height=4, width=40, wrap=tk.WORD)
                widget.bind('<KeyRelease>', self.on_field_change)
            else:
                widget = ttk.Entry(self.scrollable_frame, width=40)
                widget.bind('<KeyRelease>', self.on_field_change)
            
            widget.grid(row=row, column=1, sticky='ew', padx=5, pady=2)
            self.field_widgets[field] = widget
            row += 1
        
        # Add separator
        ttk.Separator(self.scrollable_frame, orient='horizontal').grid(row=row, column=0, columnspan=2, sticky='ew', padx=5, pady=10)
        row += 1
        
        # Tags section
        ttk.Label(self.scrollable_frame, text="Tags:", font=('TkDefaultFont', 10, 'bold')).grid(row=row, column=0, sticky='nw', padx=5, pady=2)
        
        tags_help = ttk.Label(self.scrollable_frame, text="Enter tags separated by commas", font=('TkDefaultFont', 8))
        tags_help.grid(row=row, column=1, sticky='w', padx=5, pady=2)
        row += 1
        
        self.tags_text = tk.Text(self.scrollable_frame, height=3, width=40, wrap=tk.WORD)
        self.tags_text.bind('<KeyRelease>', self.on_field_change)
        self.tags_text.grid(row=row, column=0, columnspan=2, sticky='ew', padx=5, pady=2)
        row += 1
        
        # Notes section
        ttk.Label(self.scrollable_frame, text="Notes:", font=('TkDefaultFont', 10, 'bold')).grid(row=row, column=0, sticky='nw', padx=5, pady=2)
        
        notes_help = ttk.Label(self.scrollable_frame, text="Enter notes separated by semicolons", font=('TkDefaultFont', 8))
        notes_help.grid(row=row, column=1, sticky='w', padx=5, pady=2)
        row += 1
        
        self.notes_text = tk.Text(self.scrollable_frame, height=4, width=40, wrap=tk.WORD)
        self.notes_text.bind('<KeyRelease>', self.on_field_change)
        self.notes_text.grid(row=row, column=0, columnspan=2, sticky='ew', padx=5, pady=2)
        
        # Configure grid weights
        self.scrollable_frame.columnconfigure(1, weight=1)
    
    
    def create_json_field(self):
        """Create raw JSON editor"""
        ttk.Label(self.json_frame, text="Raw JSON:", font=('TkDefaultFont', 10, 'bold')).pack(anchor=tk.W, padx=5, pady=5)
        
        json_help = ttk.Label(self.json_frame, text="Edit the raw JSON data (be careful with syntax!)", 
                             font=('TkDefaultFont', 8))
        json_help.pack(anchor=tk.W, padx=5)
        
        self.json_text = scrolledtext.ScrolledText(self.json_frame, wrap=tk.NONE, font=('Courier', 10))
        self.json_text.pack(fill=tk.BOTH, expand=True, padx=5, pady=5)
        self.json_text.bind('<KeyRelease>', self.on_field_change)
    
    def add_new_field_dialog(self):
        """Dialog to add a new field to all items"""
        dialog = tk.Toplevel(self.root)
        dialog.title("Add New Field to All Items")
        dialog.geometry("400x300")
        dialog.transient(self.root)
        dialog.grab_set()
        
        # Field name
        ttk.Label(dialog, text="Field Name:").pack(pady=5)
        name_entry = ttk.Entry(dialog, width=30)
        name_entry.pack(pady=5)
        
        # Default value
        ttk.Label(dialog, text="Default Value:").pack(pady=(10, 5))
        value_text = tk.Text(dialog, height=4, width=40)
        value_text.pack(pady=5, padx=10, fill=tk.X)
        
        # Value type
        ttk.Label(dialog, text="Value Type:").pack(pady=(10, 5))
        type_var = tk.StringVar(value="string")
        type_frame = ttk.Frame(dialog)
        type_frame.pack(pady=5)
        
        ttk.Radiobutton(type_frame, text="String", variable=type_var, value="string").pack(side=tk.LEFT, padx=5)
        ttk.Radiobutton(type_frame, text="Number", variable=type_var, value="number").pack(side=tk.LEFT, padx=5)
        ttk.Radiobutton(type_frame, text="Boolean", variable=type_var, value="boolean").pack(side=tk.LEFT, padx=5)
        ttk.Radiobutton(type_frame, text="List", variable=type_var, value="list").pack(side=tk.LEFT, padx=5)
        ttk.Radiobutton(type_frame, text="Null", variable=type_var, value="null").pack(side=tk.LEFT, padx=5)
        
        def add_field():
            field_name = name_entry.get().strip()
            if not field_name:
                messagebox.showerror("Error", "Field name is required")
                return
            
            # Check if field already exists
            if any(field_name in item for item in self.items.values()):
                if not messagebox.askyesno("Field Exists", f"Field '{field_name}' already exists in some items. Continue?"):
                    return
            
            # Get default value
            value_str = value_text.get(1.0, tk.END).strip()
            value_type = type_var.get()
            
            # Convert value based on type
            try:
                if value_type == "string":
                    default_value = value_str
                elif value_type == "number":
                    default_value = float(value_str) if '.' in value_str else int(value_str)
                elif value_type == "boolean":
                    default_value = value_str.lower() in ('true', '1', 'yes', 'on')
                elif value_type == "list":
                    default_value = [item.strip() for item in value_str.split(',') if item.strip()] if value_str else []
                elif value_type == "null":
                    default_value = None
                else:
                    default_value = value_str
            except ValueError:
                messagebox.showerror("Error", f"Invalid value for type {value_type}")
                return
            
            # Add field to all items
            for item in self.items.values():
                if field_name not in item:
                    item[field_name] = default_value
            
            # Rebuild UI
            self.rebuild_basic_fields()
            self.refresh_item_list()
            
            messagebox.showinfo("Success", f"Added field '{field_name}' to all items")
            dialog.destroy()
        
        # Buttons
        button_frame = ttk.Frame(dialog)
        button_frame.pack(pady=10)
        
        ttk.Button(button_frame, text="Add Field", command=add_field).pack(side=tk.LEFT, padx=5)
        ttk.Button(button_frame, text="Cancel", command=dialog.destroy).pack(side=tk.LEFT, padx=5)
    
    def remove_field_dialog(self):
        """Dialog to remove a field from all items"""
        # Get all existing fields
        all_fields = set()
        for item in self.items.values():
            all_fields.update(item.keys())
        
        # Core fields that shouldn't be removed
        core_fields = {'name', 'description', 'price', 'type', 'weight', 'stack', 'ac', 'damage', 'heal', 'tags', 'notes'}
        removable_fields = sorted(all_fields - core_fields)
        
        if not removable_fields:
            messagebox.showinfo("No Fields", "No removable fields found. Core fields (name, description, etc.) cannot be removed.")
            return
        
        dialog = tk.Toplevel(self.root)
        dialog.title("Remove Field from All Items")
        dialog.geometry("400x300")
        dialog.transient(self.root)
        dialog.grab_set()
        
        ttk.Label(dialog, text="Select field to remove from all items:", font=('TkDefaultFont', 10, 'bold')).pack(pady=10)
        
        # Warning label
        warning_label = ttk.Label(dialog, text="⚠️ This action cannot be undone!", 
                                 font=('TkDefaultFont', 9), foreground='red')
        warning_label.pack(pady=5)
        
        # Field selection
        field_frame = ttk.Frame(dialog)
        field_frame.pack(fill=tk.BOTH, expand=True, padx=20, pady=10)
        
        # Scrollable listbox for fields
        listbox_frame = ttk.Frame(field_frame)
        listbox_frame.pack(fill=tk.BOTH, expand=True)
        
        scrollbar = ttk.Scrollbar(listbox_frame)
        scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        
        field_listbox = tk.Listbox(listbox_frame, yscrollcommand=scrollbar.set, selectmode=tk.MULTIPLE)
        field_listbox.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        scrollbar.config(command=field_listbox.yview)
        
        # Populate listbox
        for field in removable_fields:
            # Show field name and how many items have it
            count = sum(1 for item in self.items.values() if field in item)
            field_listbox.insert(tk.END, f"{field} (in {count}/{len(self.items)} items)")
        
        def remove_fields():
            selected_indices = field_listbox.curselection()
            if not selected_indices:
                messagebox.showerror("No Selection", "Please select at least one field to remove")
                return
            
            selected_fields = [removable_fields[i] for i in selected_indices]
            
            # Confirm removal
            field_list = ", ".join(selected_fields)
            if not messagebox.askyesno("Confirm Removal", 
                                     f"Are you sure you want to remove these fields from ALL items?\n\n{field_list}\n\nThis action cannot be undone!"):
                return
            
            # Remove fields from all items
            removed_count = 0
            for field in selected_fields:
                for item in self.items.values():
                    if field in item:
                        del item[field]
                        removed_count += 1
            
            # Rebuild UI
            self.rebuild_basic_fields()
            self.refresh_item_list()
            
            messagebox.showinfo("Success", f"Removed {len(selected_fields)} field(s) from items")
            dialog.destroy()
        
        # Buttons
        button_frame = ttk.Frame(dialog)
        button_frame.pack(pady=10)
        
        ttk.Button(button_frame, text="Remove Selected Fields", command=remove_fields).pack(side=tk.LEFT, padx=5)
        ttk.Button(button_frame, text="Cancel", command=dialog.destroy).pack(side=tk.LEFT, padx=5)
    
    def bulk_edit_dialog(self):
        """Dialog for bulk editing selected items"""
        selected_keys = self.get_selected_item_keys()
        if not selected_keys:
            messagebox.showwarning("No Selection", "Please select items to bulk edit")
            return
        
        dialog = tk.Toplevel(self.root)
        dialog.title(f"Bulk Edit {len(selected_keys)} Items")
        dialog.geometry("500x400")
        dialog.transient(self.root)
        dialog.grab_set()
        
        # Get common fields
        all_fields = set()
        for key in selected_keys:
            all_fields.update(self.items[key].keys())
        
        ttk.Label(dialog, text=f"Editing {len(selected_keys)} selected items:", font=('TkDefaultFont', 10, 'bold')).pack(pady=5)
        
        # Create field editor
        canvas = tk.Canvas(dialog)
        scrollbar = ttk.Scrollbar(dialog, orient="vertical", command=canvas.yview)
        scrollable_frame = ttk.Frame(canvas)
        
        scrollable_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        
        canvas.create_window((0, 0), window=scrollable_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        bulk_fields = {}
        bulk_checkboxes = {}
        
        row = 0
        for field in sorted(all_fields):
            if field in ['tags', 'notes']:
                continue
                
            # Checkbox to enable editing this field
            var = tk.BooleanVar()
            cb = ttk.Checkbutton(scrollable_frame, text=f"Edit {field.title()}", variable=var)
            cb.grid(row=row, column=0, sticky='w', padx=5, pady=2)
            bulk_checkboxes[field] = var
            
            # Field editor
            if field == 'description':
                widget = tk.Text(scrollable_frame, height=2, width=30, wrap=tk.WORD)
            else:
                widget = ttk.Entry(scrollable_frame, width=30)
            widget.grid(row=row, column=1, sticky='ew', padx=5, pady=2)
            bulk_fields[field] = widget
            row += 1
        
        scrollable_frame.columnconfigure(1, weight=1)
        
        canvas.pack(side="left", fill="both", expand=True, padx=10, pady=10)
        scrollbar.pack(side="right", fill="y", pady=10)
        
        def apply_bulk_edit():
            changes_made = 0
            for field, checkbox_var in bulk_checkboxes.items():
                if checkbox_var.get():  # Field is checked for editing
                    widget = bulk_fields[field]
                    if isinstance(widget, tk.Text):
                        new_value = widget.get(1.0, tk.END).strip()
                    else:
                        new_value = widget.get().strip()
                    
                    # Apply to all selected items
                    for key in selected_keys:
                        if field == 'price' and new_value:
                            try:
                                # Convert gp to cp for price field
                                self.items[key][field] = self.gp_to_cp(float(new_value))
                            except ValueError:
                                continue
                        elif field in ['weight', 'stack'] and new_value:
                            try:
                                self.items[key][field] = float(new_value) if field == 'weight' else int(new_value)
                            except ValueError:
                                continue
                        else:
                            # Handle None/null values
                            if new_value.lower() in ('none', 'null', ''):
                                self.items[key][field] = None
                            else:
                                self.items[key][field] = new_value
                    changes_made += 1
            
            if changes_made > 0:
                self.refresh_item_list()
                messagebox.showinfo("Success", f"Applied {changes_made} field changes to {len(selected_keys)} items")
            
            dialog.destroy()
        
        # Buttons
        button_frame = ttk.Frame(dialog)
        button_frame.pack(side=tk.BOTTOM, pady=10)
        
        ttk.Button(button_frame, text="Apply Changes", command=apply_bulk_edit).pack(side=tk.LEFT, padx=5)
        ttk.Button(button_frame, text="Cancel", command=dialog.destroy).pack(side=tk.LEFT, padx=5)
    
    def get_selected_item_keys(self):
        """Get keys of currently selected items"""
        selected_items = self.item_tree.selection()
        keys = []
        for item_id in selected_items:
            tags = self.item_tree.item(item_id, 'tags')
            if tags:
                keys.append(tags[0])
        return keys
    
    def select_all(self):
        """Select all visible items"""
        children = self.item_tree.get_children()
        self.item_tree.selection_set(children)
        self.update_selection_label()
    
    def clear_selection(self):
        """Clear item selection"""
        self.item_tree.selection_remove(self.item_tree.selection())
        self.update_selection_label()
    
    def update_selection_label(self):
        """Update the selection count label"""
        selected_count = len(self.item_tree.selection())
        if selected_count == 0:
            self.selection_label.config(text="No items selected")
        elif selected_count == 1:
            self.selection_label.config(text="1 item selected")
        else:
            self.selection_label.config(text=f"{selected_count} items selected")
    
    def save_all_changes(self):
        """Save all items to disk"""
        count = self.save_all_items()
        messagebox.showinfo("Save Complete", f"Saved {count} items to disk")
    
    def populate_item_list(self):
        """Populate the item list and type/tag filters"""
        # Get all types for filter
        types = set()
        tags = set()
        for item in self.items.values():
            types.add(item.get('type', 'Unknown'))
            # Add all tags from this item
            if item.get('tags'):
                tags.update(item['tags'])
        
        self.type_combo['values'] = ['All'] + sorted(list(types))
        self.type_combo.set('All')
        
        self.tag_combo['values'] = ['All'] + sorted(list(tags))
        self.tag_combo.set('All')
        
        self.refresh_item_list()
        self.status_var.set(f"Loaded {len(self.items)} items")
    
    def refresh_item_list(self):
        """Refresh the item list based on current filters and sorting"""
        # Clear existing items
        for item in self.item_tree.get_children():
            self.item_tree.delete(item)
        
        search_query = self.search_var.get().lower()
        type_filter = self.type_var.get()
        tag_filter = self.tag_var.get()
        
        filtered_items = {}
        
        for key, item in self.items.items():
            # Apply type filter
            if type_filter != 'All' and item.get('type', 'Unknown') != type_filter:
                continue
            
            # Apply tag filter
            if tag_filter != 'All':
                item_tags = item.get('tags', [])
                if tag_filter not in item_tags:
                    continue
            
            # Apply search filter
            if search_query:
                if not self.item_matches_search(item, search_query):
                    continue
            
            filtered_items[key] = item
        
        # Sort items
        if self.sort_column:
            if self.sort_column == '#0':  # Name column
                sorted_items = sorted(filtered_items.items(), 
                                     key=lambda x: x[1].get('name', x[0]).lower(), 
                                     reverse=self.sort_reverse)
            elif self.sort_column == 'Price':
                sorted_items = sorted(filtered_items.items(), 
                                     key=lambda x: x[1].get('price', 0), 
                                     reverse=self.sort_reverse)
            elif self.sort_column == 'Weight':
                sorted_items = sorted(filtered_items.items(), 
                                     key=lambda x: x[1].get('weight', 0), 
                                     reverse=self.sort_reverse)
            elif self.sort_column == 'Stack':
                sorted_items = sorted(filtered_items.items(), 
                                     key=lambda x: x[1].get('stack', 1), 
                                     reverse=self.sort_reverse)
            elif self.sort_column == 'Type':
                sorted_items = sorted(filtered_items.items(), 
                                     key=lambda x: x[1].get('type', 'Unknown').lower(), 
                                     reverse=self.sort_reverse)
            else:
                sorted_items = sorted(filtered_items.items(), key=lambda x: x[1].get('name', x[0]))
        else:
            sorted_items = sorted(filtered_items.items(), key=lambda x: x[1].get('name', x[0]))
        
        # Add to treeview
        for key, item in sorted_items:
            name = item.get('name', key)
            item_type = item.get('type', 'Unknown')
            price_gp = f"{self.cp_to_gp(item.get('price', 0)):.2f}"
            weight = f"{item.get('weight', 0)}"
            stack = f"{item.get('stack', 1)}"
            
            self.item_tree.insert('', tk.END, values=(item_type, price_gp, weight, stack), text=name, tags=(key,))
        
        self.status_var.set(f"Showing {len(filtered_items)} of {len(self.items)} items")
        self.update_selection_label()
    
    def item_matches_search(self, item, query):
        """Check if item matches search query"""
        # Search in name
        if query in item.get('name', '').lower():
            return True
        
        # Search in type
        if query in item.get('type', '').lower():
            return True
        
        # Search in description
        if query in item.get('description', '').lower():
            return True
        
        # Search in tags
        if item.get('tags'):
            if any(query in tag.lower() for tag in item['tags']):
                return True
        
        # Search in notes
        if item.get('notes'):
            if any(query in note.lower() for note in item['notes']):
                return True
        
        return False
    
    def on_search(self, *args):
        """Handle search/filter changes"""
        self.refresh_item_list()
    
    def on_item_select(self, event):
        """Handle item selection"""
        self.update_selection_label()
        
        selection = self.item_tree.selection()
        if len(selection) == 1:  # Only load for single selection
            item_id = selection[0]
            tags = self.item_tree.item(item_id, 'tags')
            if tags:
                key = tags[0]
                if key in self.items:
                    self.load_item_for_editing(key, self.items[key])
    
    def load_item_for_editing(self, key, item):
        """Load item data into editor fields"""
        self.current_item_key = key
        self.current_item_data = item.copy()
        
        # Load basic fields
        for field, widget in self.field_widgets.items():
            if isinstance(widget, tk.Text):
                widget.delete(1.0, tk.END)
                value = item.get(field, '')
                if value is not None:
                    widget.insert(1.0, str(value))
            else:
                widget.delete(0, tk.END)
                value = item.get(field, '')
                if field == 'price':
                    # Convert cp to gp for display
                    value = str(self.cp_to_gp(value)) if value is not None else ''
                elif value is not None:
                    value = str(value)
                else:
                    value = ''
                widget.insert(0, value)
        
        # Load tags
        self.tags_text.delete(1.0, tk.END)
        if item.get('tags'):
            self.tags_text.insert(1.0, ', '.join(item['tags']))
        
        # Load notes
        self.notes_text.delete(1.0, tk.END)
        if item.get('notes'):
            self.notes_text.insert(1.0, '; '.join(item['notes']))
        
        # Load image
        self.load_item_image(item.get('img'))
        
        # Load raw JSON
        self.json_text.delete(1.0, tk.END)
        self.json_text.insert(1.0, json.dumps(item, indent=2))
        
        self.save_button.configure(state='normal')
    
    def load_item_image(self, img_path):
        """Load and display item image"""
        if not img_path:
            self.image_label.config(image='', text="No image\navailable")
            return
        
        # Construct full image path
        script_dir = Path(__file__).parent
        project_root = script_dir.parent
        full_img_path = project_root / "www" / "static" / "img" / img_path
        
        try:
            if full_img_path.exists():
                # Load and resize image
                with Image.open(full_img_path) as img:
                    # Resize to fit in the display area (max 200x200)
                    img.thumbnail((200, 200), Image.Resampling.LANCZOS)
                    
                    # Convert to PhotoImage for tkinter
                    photo = ImageTk.PhotoImage(img)
                    
                    # Keep a reference to prevent garbage collection
                    self.image_label.image = photo
                    self.image_label.config(image=photo, text="")
            else:
                self.image_label.config(image='', text=f"Image not found:\n{img_path}")
        except Exception as e:
            self.image_label.config(image='', text=f"Error loading image:\n{str(e)}")
    
    def on_field_change(self, event=None):
        """Handle field changes"""
        if self.current_item_key:
            self.save_button.configure(state='normal')
    
    def save_current_item(self):
        """Save the currently edited item"""
        if not self.current_item_key:
            return
        
        try:
            # Get current tab
            current_tab = self.notebook.select()
            tab_text = self.notebook.tab(current_tab, "text")
            
            if tab_text == "Raw JSON":
                # Parse from JSON tab
                json_content = self.json_text.get(1.0, tk.END).strip()
                item_data = json.loads(json_content)
            else:
                # Build from form fields
                item_data = {}
                
                # Basic fields
                for field, widget in self.field_widgets.items():
                    if isinstance(widget, tk.Text):
                        value = widget.get(1.0, tk.END).strip()
                    else:
                        value = widget.get().strip()
                    
                    # Handle special fields
                    if field == 'price':
                        item_data[field] = self.gp_to_cp(value) if value else 0
                    elif field in ['weight', 'stack']:
                        if value:
                            item_data[field] = float(value) if field == 'weight' else int(value)
                        else:
                            item_data[field] = 0 if field == 'weight' else 1
                    else:
                        # Handle None/empty values
                        if value.lower() in ('none', 'null', '') and field in ['ac', 'damage', 'heal']:
                            item_data[field] = None
                        else:
                            item_data[field] = value if value else ''
                
                # Tags
                tags_text = self.tags_text.get(1.0, tk.END).strip()
                if tags_text:
                    item_data['tags'] = [tag.strip() for tag in tags_text.split(',') if tag.strip()]
                else:
                    item_data['tags'] = []
                
                # Notes
                notes_text = self.notes_text.get(1.0, tk.END).strip()
                if notes_text:
                    item_data['notes'] = [note.strip() for note in notes_text.split(';') if note.strip()]
                else:
                    item_data['notes'] = []
            
            # Save to file
            if self.save_item(self.current_item_key, item_data):
                # Update in-memory data
                self.items[self.current_item_key] = item_data
                self.current_item_data = item_data.copy()
                
                # Update JSON tab if we're not currently on it
                if tab_text != "Raw JSON":
                    self.json_text.delete(1.0, tk.END)
                    self.json_text.insert(1.0, json.dumps(item_data, indent=2))
                
                # Update image if img field changed
                if 'img' in item_data:
                    self.load_item_image(item_data['img'])
                
                # Refresh the item list
                self.refresh_item_list()
                
                self.status_var.set(f"Saved {item_data.get('name', self.current_item_key)}")
                messagebox.showinfo("Success", "Item saved successfully!")
            
        except json.JSONDecodeError as e:
            messagebox.showerror("JSON Error", f"Invalid JSON format: {e}")
        except ValueError as e:
            messagebox.showerror("Value Error", f"Invalid value in field: {e}")
        except Exception as e:
            messagebox.showerror("Save Error", f"Failed to save item: {e}")

def main():
    root = tk.Tk()
    app = AdvancedItemEditor(root)
    root.mainloop()

if __name__ == "__main__":
    main()