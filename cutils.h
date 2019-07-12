#ifndef _CUTILS_H
#define _CUTILS_H
/* Method declaration */

/* Return the content by a given file (using fread)*/
char *read_content(const char *filename);
void read_content_no_alloc(const char *filename, char *out);

long get_file_size(char *filename);

/* Verify that 'to_find' is present in the given data */
int verify_presence_data(const char *data, const char *to_find);

/* Wrap the read_content method and verify if 'to_find' is in the content of the
 * given file */
int verify_presence_filename(const char *filename, const char *to_find);

/* Set the the file size in byte into the fsize pointer in input */
void set_file_size(FILE *fp, int *fsize);

/* Lowerize the given string */
void lowerize_string(char *data);

/* Lowerize the input 'data' and verify if 'to_find' is present */
int verify_presence_data_insensitive(char *data, char *to_find);

#endif
