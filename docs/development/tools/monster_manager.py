#!/usr/bin/env python3
"""
Monster Manager GUI - Advanced tool for managing monster JSON files
"""

import json
import os
import sys
import re
import tkinter as tk
from tkinter import ttk, messagebox, simpledialog
from pathlib import Path

MONSTER_DIR = Path("www/data/monster")

def sanitize_filename(name):
    """Convert monster name to safe filename"""
    name = re.sub(r'[^\w\s-]', '', name)
    name = re.sub(r'\s+', '-', name)
    return name.lower()

def get_monster_path(name):
    """Get the file path for a monster by name"""
    filename = sanitize_filename(name) + '.json'
    return MONSTER_DIR / filename

class MonsterManagerGUI:
    def __init__(self, root):
        self.root = root
        self.root.title("Monster Manager - Advanced")
        self.root.geometry("1400x800")

        self.selected_monsters = []
        self.monsters_data = []
        self.sort_column = 'challenge_rating'
        self.sort_reverse = False
        self.filter_criteria = {}

        self.create_widgets()
        self.refresh_monster_list()

    def create_widgets(self):
        # Main container
        main_frame = ttk.Frame(self.root, padding="10")
        main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))

        # Configure grid weights
        self.root.columnconfigure(0, weight=1)
        self.root.rowconfigure(0, weight=1)
        main_frame.columnconfigure(1, weight=1)
        main_frame.rowconfigure(2, weight=1)

        # Toolbar
        self.create_toolbar(main_frame)

        # Search and filter frame
        self.create_search_filter_frame(main_frame)

        # Monster list frame
        self.create_monster_list_frame(main_frame)

        # Details frame
        self.create_details_frame(main_frame)

    def create_toolbar(self, parent):
        toolbar = ttk.Frame(parent)
        toolbar.grid(row=0, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 10))

        # Single monster operations
        single_frame = ttk.LabelFrame(toolbar, text="Single Monster", padding="5")
        single_frame.pack(side=tk.LEFT, padx=(0, 10))

        ttk.Button(single_frame, text="Add Monster", command=self.add_monster_dialog).pack(side=tk.LEFT, padx=2)
        ttk.Button(single_frame, text="Edit Monster", command=self.edit_monster_dialog).pack(side=tk.LEFT, padx=2)
        ttk.Button(single_frame, text="Remove Monster", command=self.remove_monster_dialog).pack(side=tk.LEFT, padx=2)

        # Bulk operations
        bulk_frame = ttk.LabelFrame(toolbar, text="Bulk Operations", padding="5")
        bulk_frame.pack(side=tk.LEFT, padx=(0, 10))

        ttk.Button(bulk_frame, text="Bulk Edit", command=self.bulk_edit_dialog).pack(side=tk.LEFT, padx=2)
        ttk.Button(bulk_frame, text="Add Field", command=self.bulk_add_field_dialog).pack(side=tk.LEFT, padx=2)
        ttk.Button(bulk_frame, text="Remove Field", command=self.bulk_remove_field_dialog).pack(side=tk.LEFT, padx=2)

        # Utilities
        util_frame = ttk.LabelFrame(toolbar, text="Utilities", padding="5")
        util_frame.pack(side=tk.LEFT, padx=(0, 10))

        ttk.Button(util_frame, text="Advanced Filter", command=self.advanced_filter_dialog).pack(side=tk.LEFT, padx=2)
        ttk.Button(util_frame, text="Export Selection", command=self.export_selection).pack(side=tk.LEFT, padx=2)
        ttk.Button(util_frame, text="Refresh", command=self.refresh_monster_list).pack(side=tk.LEFT, padx=2)

    def create_search_filter_frame(self, parent):
        search_frame = ttk.Frame(parent)
        search_frame.grid(row=1, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 10))
        search_frame.columnconfigure(1, weight=1)

        ttk.Label(search_frame, text="Quick Search:").grid(row=0, column=0, padx=(0, 5))
        self.search_var = tk.StringVar()
        self.search_var.trace('w', self.on_search_change)
        search_entry = ttk.Entry(search_frame, textvariable=self.search_var)
        search_entry.grid(row=0, column=1, sticky=(tk.W, tk.E), padx=(0, 10))

        ttk.Button(search_frame, text="Clear", command=self.clear_search).grid(row=0, column=2, padx=(0, 10))

        # Filter status
        self.filter_status_var = tk.StringVar()
        self.filter_status_label = ttk.Label(search_frame, textvariable=self.filter_status_var, foreground="blue")
        self.filter_status_label.grid(row=0, column=3, padx=(0, 10))

        # Selection info
        self.selection_info_var = tk.StringVar()
        self.selection_info_label = ttk.Label(search_frame, textvariable=self.selection_info_var, foreground="green")
        self.selection_info_label.grid(row=0, column=4)

    def create_monster_list_frame(self, parent):
        list_frame = ttk.Frame(parent)
        list_frame.grid(row=2, column=0, sticky=(tk.W, tk.E, tk.N, tk.S), padx=(0, 10))
        list_frame.columnconfigure(0, weight=1)
        list_frame.rowconfigure(1, weight=1)

        ttk.Label(list_frame, text="Monsters", font=('TkDefaultFont', 10, 'bold')).grid(row=0, column=0, sticky=tk.W)

        # Treeview for monster list
        columns = ('name', 'cr', 'type', 'size', 'ac', 'hp', 'alignment', 'tags')
        self.monster_tree = ttk.Treeview(list_frame, columns=columns, show='headings', height=20, selectmode='extended')

        # Define headings - make them clickable for sorting
        headers = {
            'name': 'Name',
            'cr': 'CR',
            'type': 'Type',
            'size': 'Size',
            'ac': 'AC',
            'hp': 'HP',
            'alignment': 'Alignment',
            'tags': 'Tags'
        }

        for col, header in headers.items():
            self.monster_tree.heading(col, text=header, command=lambda c=col: self.sort_by_column(c))

        # Define column widths
        widths = {
            'name': 200,
            'cr': 60,
            'type': 120,
            'size': 80,
            'ac': 50,
            'hp': 60,
            'alignment': 120,
            'tags': 180
        }

        for col, width in widths.items():
            self.monster_tree.column(col, width=width, minwidth=width)

        self.monster_tree.grid(row=1, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        self.monster_tree.bind('<<TreeviewSelect>>', self.on_monster_select)

        # Scrollbar for treeview
        tree_scroll = ttk.Scrollbar(list_frame, orient=tk.VERTICAL, command=self.monster_tree.yview)
        tree_scroll.grid(row=1, column=1, sticky=(tk.N, tk.S))
        self.monster_tree.configure(yscrollcommand=tree_scroll.set)

        # Selection controls
        select_frame = ttk.Frame(list_frame)
        select_frame.grid(row=2, column=0, sticky=(tk.W, tk.E), pady=(5, 0))

        ttk.Button(select_frame, text="Select All", command=self.select_all).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(select_frame, text="Select None", command=self.select_none).pack(side=tk.LEFT, padx=(0, 5))
        ttk.Button(select_frame, text="Invert Selection", command=self.invert_selection).pack(side=tk.LEFT)

    def create_details_frame(self, parent):
        details_frame = ttk.LabelFrame(parent, text="Monster Details", padding="10")
        details_frame.grid(row=2, column=1, sticky=(tk.W, tk.E, tk.N, tk.S))
        details_frame.columnconfigure(0, weight=1)
        details_frame.rowconfigure(0, weight=1)

        # Create notebook for tabs
        self.notebook = ttk.Notebook(details_frame)
        self.notebook.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))

        # Field Editor Tab (first tab)
        self.editor_frame = ttk.Frame(self.notebook)
        self.notebook.add(self.editor_frame, text="Field Editor")

        # JSON Tab (second tab)
        self.json_frame = ttk.Frame(self.notebook)
        self.notebook.add(self.json_frame, text="Raw JSON")

        self.json_text = tk.Text(self.json_frame, width=40, height=20, wrap=tk.WORD, font=('Consolas', 10))
        self.json_text.pack(fill=tk.BOTH, expand=True, side=tk.LEFT)

        json_scroll = ttk.Scrollbar(self.json_frame, orient=tk.VERTICAL, command=self.json_text.yview)
        json_scroll.pack(side=tk.RIGHT, fill=tk.Y)
        self.json_text.configure(yscrollcommand=json_scroll.set)

        # Create scrollable frame for field editor
        self.editor_canvas = tk.Canvas(self.editor_frame)
        self.editor_scrollbar = ttk.Scrollbar(self.editor_frame, orient="vertical", command=self.editor_canvas.yview)
        self.editor_content = ttk.Frame(self.editor_canvas)

        self.editor_content.bind(
            "<Configure>",
            lambda e: self.editor_canvas.configure(scrollregion=self.editor_canvas.bbox("all"))
        )

        self.editor_canvas.create_window((0, 0), window=self.editor_content, anchor="nw")
        self.editor_canvas.configure(yscrollcommand=self.editor_scrollbar.set)

        self.editor_canvas.pack(side="left", fill="both", expand=True)
        self.editor_scrollbar.pack(side="right", fill="y")

        # Bind mousewheel to canvas
        self.editor_canvas.bind("<MouseWheel>", self._on_mousewheel)

        # Field editor variables
        self.field_vars = {}
        self.current_monster = None

        # Save button for field editor
        self.save_button = ttk.Button(self.editor_frame, text="Save Changes", command=self.save_field_changes)
        self.save_button.pack(side=tk.BOTTOM, pady=5)

        # Add initial empty label to prevent initialization issues
        ttk.Label(self.editor_content, text="Select a monster to edit").pack(pady=20)

    def _on_mousewheel(self, event):
        self.editor_canvas.yview_scroll(int(-1*(event.delta/120)), "units")

    def load_monsters(self):
        """Load all monsters from JSON files"""
        if not MONSTER_DIR.exists():
            return []

        monsters = []
        for file_path in MONSTER_DIR.glob("*.json"):
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    monster = json.load(f)
                    monster['filename'] = file_path.name
                    monsters.append(monster)
            except Exception as e:
                print(f"Error reading {file_path}: {e}")

        return monsters

    def apply_filters_and_sort(self, monsters):
        """Apply current filters and sorting to monster list"""
        # Apply filters
        filtered = monsters

        # Quick search filter
        search_query = self.search_var.get().strip().lower()
        if search_query:
            filtered = []
            for monster in monsters:
                searchable = [
                    str(monster.get(field, '')).lower()
                    for field in ['name', 'type', 'alignment']
                ]
                if 'tags' in monster:
                    searchable.extend([tag.lower() for tag in monster['tags']])

                if any(search_query in text for text in searchable):
                    filtered.append(monster)

        # Advanced filters
        for field, criteria in self.filter_criteria.items():
            if not criteria:
                continue

            if criteria['type'] == 'range':
                min_val, max_val = criteria['min'], criteria['max']
                filtered = [m for m in filtered if min_val <= m.get(field, 0) <= max_val]
            elif criteria['type'] == 'contains':
                value = criteria['value'].lower()
                filtered = [m for m in filtered if value in str(m.get(field, '')).lower()]
            elif criteria['type'] == 'equals':
                value = criteria['value']
                filtered = [m for m in filtered if m.get(field) == value]

        # Sort
        if self.sort_column and self.sort_column in ['name', 'challenge_rating', 'type', 'size', 'armor_class', 'hit_points', 'alignment']:
            def sort_key(monster):
                value = monster.get(self.sort_column, '')
                if self.sort_column == 'challenge_rating':
                    return float(value) if value != '' else 0
                elif self.sort_column in ['armor_class', 'hit_points']:
                    return int(value) if value != '' else 0
                return str(value).lower()

            filtered.sort(key=sort_key, reverse=self.sort_reverse)

        return filtered

    def refresh_monster_list(self):
        """Refresh the monster list display"""
        self.monsters_data = self.load_monsters()
        filtered_monsters = self.apply_filters_and_sort(self.monsters_data)

        # Clear existing items
        for item in self.monster_tree.get_children():
            self.monster_tree.delete(item)

        # Populate treeview
        for monster in filtered_monsters:
            tags_str = ', '.join(monster.get('tags', []))

            item_id = self.monster_tree.insert('', 'end', values=(
                monster.get('name', ''),
                monster.get('challenge_rating', ''),
                monster.get('type', ''),
                monster.get('size', ''),
                monster.get('armor_class', ''),
                monster.get('hit_points', ''),
                monster.get('alignment', ''),
                tags_str
            ))

            # Store reference to monster data for easy access later
            # We'll find monsters by name when needed

        # Update status
        total_count = len(self.monsters_data)
        filtered_count = len(filtered_monsters)
        selected_count = len(self.selected_monsters)

        if filtered_count != total_count:
            self.filter_status_var.set(f"Showing {filtered_count} of {total_count} monsters")
        else:
            self.filter_status_var.set("")

        self.selection_info_var.set(f"Selected: {selected_count}")

        # Update visual selection in treeview (only if we have selected monsters)
        if hasattr(self, 'selected_monsters') and self.selected_monsters:
            self.update_treeview_selection()

    def update_treeview_selection(self):
        """Update visual selection in treeview to match selected_monsters"""
        # Clear current selection
        self.monster_tree.selection_remove(self.monster_tree.selection())

        # Select items that match selected monsters
        selected_names = [m.get('name') for m in self.selected_monsters]
        for item in self.monster_tree.get_children():
            item_values = self.monster_tree.item(item)['values']
            if item_values and item_values[0] in selected_names:  # name is first column now
                self.monster_tree.selection_add(item)

    def sort_by_column(self, column):
        """Sort monsters by clicked column"""
        if self.sort_column == column:
            self.sort_reverse = not self.sort_reverse
        else:
            self.sort_column = column
            self.sort_reverse = False

        # Update column header to show sort direction
        for col in ['name', 'cr', 'type', 'size', 'ac', 'hp', 'alignment']:
            header_text = {
                'name': 'Name', 'cr': 'CR', 'type': 'Type', 'size': 'Size',
                'ac': 'AC', 'hp': 'HP', 'alignment': 'Alignment'
            }[col]

            if col == column:
                arrow = '↓' if self.sort_reverse else '↑'
                header_text += f' {arrow}'

            self.monster_tree.heading(col, text=header_text)

        self.refresh_monster_list()


    def select_all(self):
        """Select all visible monsters"""
        filtered_monsters = self.apply_filters_and_sort(self.monsters_data)
        self.selected_monsters = filtered_monsters.copy()
        self.refresh_monster_list()

    def select_none(self):
        """Clear all selections"""
        self.selected_monsters = []
        self.refresh_monster_list()

    def invert_selection(self):
        """Invert current selection"""
        filtered_monsters = self.apply_filters_and_sort(self.monsters_data)
        selected_names = [m.get('name') for m in self.selected_monsters]

        new_selection = []
        for monster in filtered_monsters:
            if monster.get('name') not in selected_names:
                new_selection.append(monster)

        self.selected_monsters = new_selection
        self.refresh_monster_list()

    def on_search_change(self, *args):
        """Handle search text change"""
        self.refresh_monster_list()

    def clear_search(self):
        """Clear search and filters"""
        self.search_var.set('')
        self.filter_criteria = {}
        self.refresh_monster_list()

    def on_monster_select(self, event):
        """Handle monster selection for details view and multi-selection"""
        selection = self.monster_tree.selection()
        if not selection:
            # Clear selected monsters and details
            self.selected_monsters = []
            self.current_monster = None
            # Clear field editor
            for widget in self.editor_content.winfo_children():
                widget.destroy()
            # Clear JSON
            self.json_text.delete(1.0, tk.END)
            self.selection_info_var.set("Selected: 0")
            return

        # Update selected monsters based on treeview selection
        self.selected_monsters = []
        for item_id in selection:
            item = self.monster_tree.item(item_id)
            monster_name = item['values'][0]  # name is first column now
            monster = next((m for m in self.monsters_data if m.get('name') == monster_name), None)
            if monster:
                self.selected_monsters.append(monster)

        # Update selection count
        self.selection_info_var.set(f"Selected: {len(self.selected_monsters)}")

        # Display details for the first/last selected monster
        if self.selected_monsters:
            self.display_monster_details(self.selected_monsters[-1])  # Show last selected

    def display_monster_details(self, monster):
        """Display monster details in both tabs"""
        self.current_monster = monster

        # Update JSON tab
        self.json_text.delete(1.0, tk.END)
        clean_monster = {k: v for k, v in monster.items() if k != 'filename'}
        json_str = json.dumps(clean_monster, indent=2, ensure_ascii=False)
        self.json_text.insert(1.0, json_str)

        # Update Field Editor tab
        self.update_field_editor(monster)

    def update_field_editor(self, monster):
        """Update the field editor with monster data"""
        # Clear existing widgets
        for widget in self.editor_content.winfo_children():
            widget.destroy()

        self.field_vars = {}

        # Create field editors
        row = 0
        for key, value in monster.items():
            if key == 'filename':
                continue

            # Create label
            label = ttk.Label(self.editor_content, text=f"{key.replace('_', ' ').title()}:")
            label.grid(row=row, column=0, sticky=tk.W, padx=5, pady=3)

            # Create appropriate input widget based on field type
            if key == 'tags' and isinstance(value, list):
                var = tk.StringVar(value=', '.join(value))
                widget = ttk.Entry(self.editor_content, textvariable=var, width=40)
            elif key in ['challenge_rating']:
                var = tk.StringVar(value=str(value))
                widget = ttk.Entry(self.editor_content, textvariable=var, width=40)
            elif key in ['armor_class', 'hit_points', 'id']:
                var = tk.StringVar(value=str(value))
                widget = ttk.Entry(self.editor_content, textvariable=var, width=40)
            elif key == 'size':
                var = tk.StringVar(value=str(value))
                widget = ttk.Combobox(self.editor_content, textvariable=var,
                                    values=['Tiny', 'Small', 'Medium', 'Large', 'Huge', 'Gargantuan'],
                                    width=37)
            elif key == 'alignment':
                var = tk.StringVar(value=str(value))
                widget = ttk.Combobox(self.editor_content, textvariable=var,
                                    values=['lawful good', 'neutral good', 'chaotic good',
                                          'lawful neutral', 'neutral', 'chaotic neutral',
                                          'lawful evil', 'neutral evil', 'chaotic evil', 'unaligned'],
                                    width=37)
            else:
                var = tk.StringVar(value=str(value))
                widget = ttk.Entry(self.editor_content, textvariable=var, width=40)

            widget.grid(row=row, column=1, sticky=(tk.W, tk.E), padx=5, pady=3)
            self.field_vars[key] = var

            # Add delete button for non-essential fields
            if key not in ['id', 'name', 'challenge_rating', 'type', 'size', 'armor_class', 'hit_points', 'alignment']:
                delete_btn = ttk.Button(self.editor_content, text="×", width=3,
                                      command=lambda k=key: self.delete_field(k))
                delete_btn.grid(row=row, column=2, padx=5, pady=3)

            row += 1

        # Add new field section
        ttk.Separator(self.editor_content, orient='horizontal').grid(row=row, column=0, columnspan=3,
                                                                    sticky=(tk.W, tk.E), pady=10)
        row += 1

        ttk.Label(self.editor_content, text="Add New Field:", font=('TkDefaultFont', 10, 'bold')).grid(
            row=row, column=0, columnspan=3, sticky=tk.W, padx=5, pady=5)
        row += 1

        ttk.Label(self.editor_content, text="Field Name:").grid(row=row, column=0, sticky=tk.W, padx=5, pady=3)
        self.new_field_name_var = tk.StringVar()
        ttk.Entry(self.editor_content, textvariable=self.new_field_name_var, width=20).grid(
            row=row, column=1, sticky=tk.W, padx=5, pady=3)
        row += 1

        ttk.Label(self.editor_content, text="Field Value:").grid(row=row, column=0, sticky=tk.W, padx=5, pady=3)
        self.new_field_value_var = tk.StringVar()
        ttk.Entry(self.editor_content, textvariable=self.new_field_value_var, width=40).grid(
            row=row, column=1, sticky=(tk.W, tk.E), padx=5, pady=3)

        ttk.Button(self.editor_content, text="Add Field", command=self.add_new_field).grid(
            row=row, column=2, padx=5, pady=3)

        # Configure column weights
        self.editor_content.columnconfigure(1, weight=1)

    def delete_field(self, field_name):
        """Delete a field from the current monster"""
        if messagebox.askyesno("Confirm Delete", f"Delete field '{field_name}'?"):
            if field_name in self.field_vars:
                del self.field_vars[field_name]
            if field_name in self.current_monster:
                del self.current_monster[field_name]
            self.update_field_editor(self.current_monster)

    def add_new_field(self):
        """Add a new field to the current monster"""
        field_name = self.new_field_name_var.get().strip()
        field_value = self.new_field_value_var.get().strip()

        if not field_name:
            messagebox.showerror("Error", "Field name is required!")
            return

        # Convert field name to lowercase with underscores
        field_name = re.sub(r'[^\w]', '_', field_name.lower())

        # Try to convert value to appropriate type
        if field_value.replace('.', '').replace('-', '').isdigit():
            if '.' in field_value:
                field_value = float(field_value)
            else:
                field_value = int(field_value)

        self.current_monster[field_name] = field_value
        self.new_field_name_var.set('')
        self.new_field_value_var.set('')
        self.update_field_editor(self.current_monster)

    def save_field_changes(self):
        """Save changes from field editor"""
        if not self.current_monster:
            messagebox.showwarning("No Monster", "No monster selected to save!")
            return

        try:
            # Collect changes from field editor
            for field_name, var in self.field_vars.items():
                value = var.get().strip()

                if field_name == 'tags':
                    # Handle tags specially
                    if value:
                        self.current_monster[field_name] = [tag.strip() for tag in value.split(',') if tag.strip()]
                    else:
                        if field_name in self.current_monster:
                            del self.current_monster[field_name]
                elif field_name in ['challenge_rating']:
                    self.current_monster[field_name] = float(value) if value else 0.0
                elif field_name in ['armor_class', 'hit_points', 'id']:
                    self.current_monster[field_name] = int(value) if value else 0
                else:
                    self.current_monster[field_name] = value

            # Save to file
            self.save_monster(self.current_monster)

            # Update the display
            self.refresh_monster_list()
            self.display_monster_details(self.current_monster)

            messagebox.showinfo("Success", f"Monster '{self.current_monster['name']}' saved successfully!")

        except Exception as e:
            messagebox.showerror("Error", f"Failed to save changes: {e}")

    # Dialogs and operations
    def add_monster_dialog(self):
        """Open dialog to add new monster"""
        dialog = MonsterDialog(self.root, "Add Monster")
        if dialog.result:
            self.save_monster(dialog.result)
            self.refresh_monster_list()

    def edit_monster_dialog(self):
        """Open dialog to edit selected monster"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select a monster to edit.")
            return

        if len(self.selected_monsters) > 1:
            messagebox.showwarning("Multiple Selection", "Please select only one monster to edit.")
            return

        monster = self.selected_monsters[0]
        dialog = MonsterDialog(self.root, "Edit Monster", monster)
        if dialog.result:
            # Remove old file if name changed
            old_name = monster.get('name')
            new_name = dialog.result.get('name')
            if old_name != new_name:
                old_path = get_monster_path(old_name)
                if old_path.exists():
                    old_path.unlink()

            self.save_monster(dialog.result)
            self.refresh_monster_list()

    def remove_monster_dialog(self):
        """Remove selected monsters with confirmation"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select monsters to remove.")
            return

        count = len(self.selected_monsters)
        if messagebox.askyesno("Confirm Removal", f"Are you sure you want to remove {count} monster(s)?"):
            for monster in self.selected_monsters:
                monster_path = get_monster_path(monster['name'])
                if monster_path.exists():
                    monster_path.unlink()

            messagebox.showinfo("Success", f"{count} monster(s) removed successfully!")
            self.selected_monsters = []
            self.refresh_monster_list()

    def bulk_edit_dialog(self):
        """Open bulk edit dialog"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select monsters to bulk edit.")
            return

        dialog = BulkEditDialog(self.root, self.selected_monsters)
        if dialog.result:
            self.apply_bulk_changes(dialog.result)

    def bulk_add_field_dialog(self):
        """Open dialog to add field to selected monsters"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select monsters to add field to.")
            return

        dialog = AddFieldDialog(self.root)
        if dialog.result:
            field_name, field_value = dialog.result
            for monster in self.selected_monsters:
                monster[field_name] = field_value
                self.save_monster(monster)

            messagebox.showinfo("Success", f"Field '{field_name}' added to {len(self.selected_monsters)} monsters!")
            self.refresh_monster_list()

    def bulk_remove_field_dialog(self):
        """Open dialog to remove field from selected monsters"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select monsters to remove field from.")
            return

        # Get all possible fields from selected monsters
        all_fields = set()
        for monster in self.selected_monsters:
            all_fields.update(monster.keys())

        # Remove system fields
        system_fields = ['id', 'filename']
        available_fields = sorted([f for f in all_fields if f not in system_fields])

        if not available_fields:
            messagebox.showinfo("No Fields", "No removable fields found in selected monsters.")
            return

        dialog = RemoveFieldDialog(self.root, available_fields)
        if dialog.result:
            field_name = dialog.result
            count = 0
            for monster in self.selected_monsters:
                if field_name in monster:
                    del monster[field_name]
                    self.save_monster(monster)
                    count += 1

            messagebox.showinfo("Success", f"Field '{field_name}' removed from {count} monsters!")
            self.refresh_monster_list()

    def advanced_filter_dialog(self):
        """Open advanced filter dialog"""
        dialog = AdvancedFilterDialog(self.root, self.filter_criteria)
        if dialog.result is not None:
            self.filter_criteria = dialog.result
            self.refresh_monster_list()

    def export_selection(self):
        """Export selected monsters to JSON file"""
        if not self.selected_monsters:
            messagebox.showwarning("No Selection", "Please select monsters to export.")
            return

        filename = simpledialog.askstring("Export", "Enter filename (without .json):")
        if filename:
            try:
                export_path = Path(f"{filename}.json")
                with open(export_path, 'w', encoding='utf-8') as f:
                    # Remove filename field for export
                    export_data = []
                    for monster in self.selected_monsters:
                        clean_monster = {k: v for k, v in monster.items() if k != 'filename'}
                        export_data.append(clean_monster)

                    json.dump(export_data, f, indent=2, ensure_ascii=False)

                messagebox.showinfo("Success", f"Exported {len(self.selected_monsters)} monsters to {export_path}")
            except Exception as e:
                messagebox.showerror("Error", f"Export failed: {e}")

    def apply_bulk_changes(self, changes):
        """Apply bulk changes to selected monsters"""
        count = 0
        for monster in self.selected_monsters:
            modified = False
            for field, value in changes.items():
                if value is not None and value != '':  # Only apply non-empty values
                    if field == 'tags' and isinstance(value, str):
                        # Handle tags specially
                        existing_tags = monster.get('tags', [])
                        new_tags = [tag.strip() for tag in value.split(',') if tag.strip()]
                        # Merge tags (avoid duplicates)
                        all_tags = list(set(existing_tags + new_tags))
                        if all_tags:
                            monster['tags'] = all_tags
                        modified = True
                    else:
                        monster[field] = value
                        modified = True

            if modified:
                self.save_monster(monster)
                count += 1

        messagebox.showinfo("Success", f"Bulk changes applied to {count} monsters!")
        self.refresh_monster_list()

    def save_monster(self, monster_data):
        """Save monster data to JSON file"""
        try:
            # Get next ID if adding new monster
            if 'id' not in monster_data:
                max_id = 0
                for monster in self.monsters_data:
                    max_id = max(max_id, monster.get('id', 0))
                monster_data['id'] = max_id + 1

            # Remove empty tags and filename
            clean_data = {k: v for k, v in monster_data.items() if k != 'filename'}
            if 'tags' in clean_data and not clean_data['tags']:
                del clean_data['tags']

            monster_path = get_monster_path(monster_data['name'])
            MONSTER_DIR.mkdir(parents=True, exist_ok=True)

            with open(monster_path, 'w', encoding='utf-8') as f:
                json.dump(clean_data, f, indent=2, ensure_ascii=False)

        except Exception as e:
            messagebox.showerror("Error", f"Failed to save monster: {e}")

# Dialog classes
class MonsterDialog:
    def __init__(self, parent, title, monster_data=None):
        self.result = None

        # Create dialog window
        self.dialog = tk.Toplevel(parent)
        self.dialog.title(title)
        self.dialog.geometry("400x600")
        self.dialog.transient(parent)
        self.dialog.grab_set()

        # Center the dialog
        self.dialog.geometry("+%d+%d" % (parent.winfo_rootx() + 50, parent.winfo_rooty() + 50))

        self.create_dialog_widgets(monster_data)

        # Wait for dialog to close
        self.dialog.wait_window()

    def create_dialog_widgets(self, monster_data):
        """Create dialog form widgets"""
        main_frame = ttk.Frame(self.dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)

        # Form fields
        fields = [
            ('name', 'Name:', 'entry'),
            ('challenge_rating', 'Challenge Rating:', 'entry'),
            ('type', 'Type:', 'entry'),
            ('size', 'Size:', 'combobox', ['Tiny', 'Small', 'Medium', 'Large', 'Huge', 'Gargantuan']),
            ('armor_class', 'Armor Class:', 'entry'),
            ('hit_points', 'Hit Points:', 'entry'),
            ('alignment', 'Alignment:', 'combobox', ['lawful good', 'neutral good', 'chaotic good',
                                                     'lawful neutral', 'neutral', 'chaotic neutral',
                                                     'lawful evil', 'neutral evil', 'chaotic evil', 'unaligned']),
            ('tags', 'Tags (comma-separated):', 'entry')
        ]

        self.field_vars = {}
        row = 0

        for field_data in fields:
            field_name = field_data[0]
            label_text = field_data[1]
            field_type = field_data[2]

            # Label
            ttk.Label(main_frame, text=label_text).grid(row=row, column=0, sticky=tk.W, pady=2)

            # Field
            if field_type == 'entry':
                var = tk.StringVar()
                widget = ttk.Entry(main_frame, textvariable=var, width=30)
            elif field_type == 'combobox':
                var = tk.StringVar()
                values = field_data[3] if len(field_data) > 3 else []
                widget = ttk.Combobox(main_frame, textvariable=var, values=values, width=27)

            widget.grid(row=row, column=1, sticky=(tk.W, tk.E), pady=2, padx=(10, 0))
            self.field_vars[field_name] = var

            # Set initial values if editing
            if monster_data and field_name in monster_data:
                if field_name == 'tags' and isinstance(monster_data[field_name], list):
                    var.set(', '.join(monster_data[field_name]))
                else:
                    var.set(str(monster_data[field_name]))

            row += 1

        main_frame.columnconfigure(1, weight=1)

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.grid(row=row, column=0, columnspan=2, pady=20)

        ttk.Button(button_frame, text="Save", command=self.save_monster).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Cancel", command=self.dialog.destroy).pack(side=tk.LEFT)

    def save_monster(self):
        """Validate and save monster data"""
        try:
            # Collect data
            monster_data = {}

            # Required fields
            required_fields = ['name', 'challenge_rating', 'type', 'size', 'armor_class', 'hit_points', 'alignment']
            for field in required_fields:
                value = self.field_vars[field].get().strip()
                if not value:
                    messagebox.showerror("Validation Error", f"{field.replace('_', ' ').title()} is required!")
                    return

                # Convert numeric fields
                if field in ['challenge_rating']:
                    monster_data[field] = float(value)
                elif field in ['armor_class', 'hit_points']:
                    monster_data[field] = int(value)
                else:
                    monster_data[field] = value

            # Handle tags
            tags_str = self.field_vars['tags'].get().strip()
            if tags_str:
                monster_data['tags'] = [tag.strip() for tag in tags_str.split(',') if tag.strip()]

            self.result = monster_data
            self.dialog.destroy()

        except ValueError as e:
            messagebox.showerror("Validation Error", "Please check numeric values (Challenge Rating, Armor Class, Hit Points)")
        except Exception as e:
            messagebox.showerror("Error", f"An error occurred: {e}")

class BulkEditDialog:
    def __init__(self, parent, selected_monsters):
        self.result = None
        self.selected_monsters = selected_monsters

        self.dialog = tk.Toplevel(parent)
        self.dialog.title("Bulk Edit Monsters")
        self.dialog.geometry("500x400")
        self.dialog.transient(parent)
        self.dialog.grab_set()

        self.create_widgets()
        self.dialog.wait_window()

    def create_widgets(self):
        main_frame = ttk.Frame(self.dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)

        ttk.Label(main_frame, text=f"Bulk editing {len(self.selected_monsters)} monsters",
                 font=('TkDefaultFont', 10, 'bold')).pack(pady=(0, 10))

        ttk.Label(main_frame, text="Leave fields empty to keep existing values",
                 foreground="blue").pack(pady=(0, 20))

        # Create form
        self.field_vars = {}
        fields = [
            ('challenge_rating', 'Challenge Rating:'),
            ('type', 'Type:'),
            ('size', 'Size:'),
            ('armor_class', 'Armor Class:'),
            ('hit_points', 'Hit Points:'),
            ('alignment', 'Alignment:'),
            ('tags', 'Additional Tags (comma-separated):')
        ]

        for field_name, label_text in fields:
            frame = ttk.Frame(main_frame)
            frame.pack(fill=tk.X, pady=2)

            ttk.Label(frame, text=label_text, width=20).pack(side=tk.LEFT)
            var = tk.StringVar()
            entry = ttk.Entry(frame, textvariable=var)
            entry.pack(side=tk.LEFT, fill=tk.X, expand=True, padx=(10, 0))
            self.field_vars[field_name] = var

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.pack(pady=20)

        ttk.Button(button_frame, text="Apply Changes", command=self.apply_changes).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Cancel", command=self.dialog.destroy).pack(side=tk.LEFT)

    def apply_changes(self):
        changes = {}
        for field_name, var in self.field_vars.items():
            value = var.get().strip()
            if value:
                if field_name in ['challenge_rating']:
                    try:
                        changes[field_name] = float(value)
                    except ValueError:
                        messagebox.showerror("Error", f"Invalid number for {field_name}")
                        return
                elif field_name in ['armor_class', 'hit_points']:
                    try:
                        changes[field_name] = int(value)
                    except ValueError:
                        messagebox.showerror("Error", f"Invalid number for {field_name}")
                        return
                else:
                    changes[field_name] = value

        self.result = changes
        self.dialog.destroy()

class AddFieldDialog:
    def __init__(self, parent):
        self.result = None

        self.dialog = tk.Toplevel(parent)
        self.dialog.title("Add Field to Monsters")
        self.dialog.geometry("400x200")
        self.dialog.transient(parent)
        self.dialog.grab_set()

        self.create_widgets()
        self.dialog.wait_window()

    def create_widgets(self):
        main_frame = ttk.Frame(self.dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)

        # Field name
        ttk.Label(main_frame, text="Field Name:").grid(row=0, column=0, sticky=tk.W, pady=5)
        self.field_name_var = tk.StringVar()
        ttk.Entry(main_frame, textvariable=self.field_name_var, width=30).grid(row=0, column=1, padx=(10, 0), pady=5)

        # Field value
        ttk.Label(main_frame, text="Field Value:").grid(row=1, column=0, sticky=tk.W, pady=5)
        self.field_value_var = tk.StringVar()
        ttk.Entry(main_frame, textvariable=self.field_value_var, width=30).grid(row=1, column=1, padx=(10, 0), pady=5)

        main_frame.columnconfigure(1, weight=1)

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.grid(row=2, column=0, columnspan=2, pady=20)

        ttk.Button(button_frame, text="Add Field", command=self.add_field).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Cancel", command=self.dialog.destroy).pack(side=tk.LEFT)

    def add_field(self):
        field_name = self.field_name_var.get().strip()
        field_value = self.field_value_var.get().strip()

        if not field_name:
            messagebox.showerror("Error", "Field name is required!")
            return

        # Convert field name to lowercase with underscores
        field_name = re.sub(r'[^\w]', '_', field_name.lower())

        self.result = (field_name, field_value)
        self.dialog.destroy()

class RemoveFieldDialog:
    def __init__(self, parent, available_fields):
        self.result = None

        self.dialog = tk.Toplevel(parent)
        self.dialog.title("Remove Field from Monsters")
        self.dialog.geometry("300x200")
        self.dialog.transient(parent)
        self.dialog.grab_set()

        self.available_fields = available_fields
        self.create_widgets()
        self.dialog.wait_window()

    def create_widgets(self):
        main_frame = ttk.Frame(self.dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)

        ttk.Label(main_frame, text="Select field to remove:").pack(pady=(0, 10))

        self.field_var = tk.StringVar()
        field_combo = ttk.Combobox(main_frame, textvariable=self.field_var,
                                  values=self.available_fields, state="readonly")
        field_combo.pack(fill=tk.X, pady=(0, 20))

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.pack()

        ttk.Button(button_frame, text="Remove Field", command=self.remove_field).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Cancel", command=self.dialog.destroy).pack(side=tk.LEFT)

    def remove_field(self):
        field_name = self.field_var.get()
        if not field_name:
            messagebox.showerror("Error", "Please select a field to remove!")
            return

        self.result = field_name
        self.dialog.destroy()

class AdvancedFilterDialog:
    def __init__(self, parent, current_filters):
        self.result = None
        self.current_filters = current_filters.copy()

        self.dialog = tk.Toplevel(parent)
        self.dialog.title("Advanced Filters")
        self.dialog.geometry("500x400")
        self.dialog.transient(parent)
        self.dialog.grab_set()

        self.create_widgets()
        self.dialog.wait_window()

    def create_widgets(self):
        main_frame = ttk.Frame(self.dialog, padding="20")
        main_frame.pack(fill=tk.BOTH, expand=True)

        ttk.Label(main_frame, text="Advanced Filters", font=('TkDefaultFont', 12, 'bold')).pack(pady=(0, 20))

        # Challenge Rating filter
        cr_frame = ttk.LabelFrame(main_frame, text="Challenge Rating", padding="10")
        cr_frame.pack(fill=tk.X, pady=(0, 10))

        ttk.Label(cr_frame, text="Min CR:").grid(row=0, column=0, sticky=tk.W)
        self.min_cr_var = tk.StringVar()
        ttk.Entry(cr_frame, textvariable=self.min_cr_var, width=10).grid(row=0, column=1, padx=5)

        ttk.Label(cr_frame, text="Max CR:").grid(row=0, column=2, sticky=tk.W, padx=(20, 0))
        self.max_cr_var = tk.StringVar()
        ttk.Entry(cr_frame, textvariable=self.max_cr_var, width=10).grid(row=0, column=3, padx=5)

        # Type filter
        type_frame = ttk.LabelFrame(main_frame, text="Type", padding="10")
        type_frame.pack(fill=tk.X, pady=(0, 10))

        self.type_var = tk.StringVar()
        ttk.Entry(type_frame, textvariable=self.type_var, width=30).pack(side=tk.LEFT)
        ttk.Label(type_frame, text="(contains)", foreground="gray").pack(side=tk.LEFT, padx=(10, 0))

        # Size filter
        size_frame = ttk.LabelFrame(main_frame, text="Size", padding="10")
        size_frame.pack(fill=tk.X, pady=(0, 10))

        self.size_var = tk.StringVar()
        sizes = ['', 'Tiny', 'Small', 'Medium', 'Large', 'Huge', 'Gargantuan']
        ttk.Combobox(size_frame, textvariable=self.size_var, values=sizes).pack()

        # Load current filter values
        if 'challenge_rating' in self.current_filters:
            cr_filter = self.current_filters['challenge_rating']
            if cr_filter.get('type') == 'range':
                self.min_cr_var.set(str(cr_filter.get('min', '')))
                self.max_cr_var.set(str(cr_filter.get('max', '')))

        if 'type' in self.current_filters:
            type_filter = self.current_filters['type']
            if type_filter.get('type') == 'contains':
                self.type_var.set(type_filter.get('value', ''))

        if 'size' in self.current_filters:
            size_filter = self.current_filters['size']
            if size_filter.get('type') == 'equals':
                self.size_var.set(size_filter.get('value', ''))

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.pack(pady=20)

        ttk.Button(button_frame, text="Apply Filters", command=self.apply_filters).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Clear All", command=self.clear_filters).pack(side=tk.LEFT, padx=(0, 10))
        ttk.Button(button_frame, text="Cancel", command=self.dialog.destroy).pack(side=tk.LEFT)

    def apply_filters(self):
        filters = {}

        # Challenge Rating filter
        min_cr = self.min_cr_var.get().strip()
        max_cr = self.max_cr_var.get().strip()
        if min_cr or max_cr:
            try:
                min_val = float(min_cr) if min_cr else 0
                max_val = float(max_cr) if max_cr else float('inf')
                filters['challenge_rating'] = {'type': 'range', 'min': min_val, 'max': max_val}
            except ValueError:
                messagebox.showerror("Error", "Invalid Challenge Rating values")
                return

        # Type filter
        type_val = self.type_var.get().strip()
        if type_val:
            filters['type'] = {'type': 'contains', 'value': type_val}

        # Size filter
        size_val = self.size_var.get().strip()
        if size_val:
            filters['size'] = {'type': 'equals', 'value': size_val}

        self.result = filters
        self.dialog.destroy()

    def clear_filters(self):
        self.result = {}
        self.dialog.destroy()

def main():
    root = tk.Tk()
    app = MonsterManagerGUI(root)
    root.mainloop()

if __name__ == '__main__':
    main()