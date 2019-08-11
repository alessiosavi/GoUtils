#ifndef _CUTILS_H
#define _CUTILS_H

void set_file_size(FILE *fp, int *fsize);
long get_file_size(char *filename);
char *read_content(const char *filename);
void read_content_no_alloc(char *filename, char *out);
void lowerize_string(char *data);
static char **makeCharArray(int size);
static void setArrayString(char **a, char *s, int n);
static void freeCharArray(char **a, int size);
static void printCharArray(char **a, int size);
size_t levenshtein_n(const char *a, const size_t length, const char *b,
                     const size_t bLength);
int levenshtein(const char *a, const char *b);
char **makeCharArray(int size);
void setArrayString(char **a, char *s, int n);
void freeCharArray(char **a, int size);
char *apply_levian_to_array(char **a, char *wrong, int size);
void printCharArray(char **a, int size);
char **split_data(char *str, int lenght);
char *spell_check(const char *filename, const int file_lines, char *wrong);
int verify_presence_data_insensitive(char *data, char *to_find);
int verify_presence_data(const char *data, const char *to_find);
int verify_presence_filename(const char *filename, const char *to_find);


#endif
